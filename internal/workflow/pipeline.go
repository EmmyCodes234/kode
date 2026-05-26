package workflow

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	kodecontext "github.com/kode/kode/internal/context"
	"github.com/kode/kode/internal/execution"
	"github.com/kode/kode/internal/llm"
)

func NewPipeline(config Config) *Pipeline {
	if config.MaxRetries <= 0 {
		config.MaxRetries = 3
	}
	return &Pipeline{
		config:      config,
		beforeStage: make(map[Stage]func(*State)),
		afterStage:  make(map[Stage]func(*State, error)),
	}
}

func (p *Pipeline) BeforeStage(stage Stage, fn func(*State)) {
	p.beforeStage[stage] = fn
}

func (p *Pipeline) AfterStage(stage Stage, fn func(*State, error)) {
	p.afterStage[stage] = fn
}

func (p *Pipeline) Run(ctx context.Context, task string) (*Result, error) {
	state := &State{
		CurrentStage: StagePlan,
		TaskID:       fmt.Sprintf("%d", time.Now().UnixNano()),
		Task:         task,
		StartTime:    time.Now(),
	}
	if p.config.LLMConfig != nil {
		state.ProjectRoot = "."
	}

	cfg := llm.DefaultConfig()
	if p.config.LLMConfig != nil {
		cfg = *p.config.LLMConfig
	}
	if p.config.ModelOverride != "" {
		cfg.Model = p.config.ModelOverride
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("LLM API key not configured")
	}

	var budgetTracker *llm.BudgetTracker
	if p.config.TokenBudget != nil {
		budgetTracker = llm.NewBudgetTracker(cfg.Model, *p.config.TokenBudget)
	}

	projectRoot := state.ProjectRoot
	absDir, err := filepath.Abs(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("invalid project directory: %w", err)
	}
	state.ProjectRoot = absDir

	// Stage: Plan — load context
	state.CurrentStage = StagePlan
	if fn, ok := p.beforeStage[StagePlan]; ok {
		fn(state)
	}
	contextStr := ""
	if p.config.ContextFile != "" {
		data, err := os.ReadFile(p.config.ContextFile)
		if err != nil {
			err = fmt.Errorf("context file: %w", err)
			if fn, ok := p.afterStage[StagePlan]; ok {
				fn(state, err)
			}
			return nil, err
		}
		contextStr = string(data)
	}

	if contextStr == "" && p.config.EnableContextIndex && absDir != "" {
		builder := kodecontext.NewBuilder()
		fullCtx, err := builder.BuildFullContext(absDir)
		if err != nil {
			// Non-fatal — continue without auto-context
			contextStr = ""
		} else {
			contextStr = fullCtx
		}
	}

	if fn, ok := p.afterStage[StagePlan]; ok {
		fn(state, nil)
	}

	// Stage: Generate — LLM call + parse hunks
	state.CurrentStage = StageGenerate
	if fn, ok := p.beforeStage[StageGenerate]; ok {
		fn(state)
	}
	if budgetTracker != nil && budgetTracker.IsExceeded() {
		err := fmt.Errorf("token budget exceeded before generation: %s", budgetTracker.ExceededMessage())
		state.Errors = append(state.Errors, err.Error())
		if fn, ok := p.afterStage[StageGenerate]; ok {
			fn(state, err)
		}
		return nil, err
	}

	userPrompt := llm.BuildGeneratePrompt(task, contextStr)
	client := llm.NewClient(cfg)
	resp, err := client.Chat(ctx, llm.ChatRequest{
		Model:       cfg.Model,
		Messages:    []llm.Message{{Role: llm.RoleSystem, Content: llm.SystemPrompt}, {Role: llm.RoleUser, Content: userPrompt}},
		Temperature: 0.2,
		MaxTokens:   4096,
	})
	if err != nil {
		err = fmt.Errorf("LLM call failed: %w", err)
		state.Errors = append(state.Errors, err.Error())
		if fn, ok := p.afterStage[StageGenerate]; ok {
			fn(state, err)
		}
		return nil, err
	}

	if budgetTracker != nil && resp.Usage != nil {
		budgetTracker.Track(resp.Usage.PromptTokens, resp.Usage.CompletionTokens)
	}

	content := resp.Choices[0].Message.Content
	parser := execution.NewHunkParser()
	hunks, err := parser.ParseLLMResponse(content)
	if err != nil {
		err = fmt.Errorf("parse error: %w", err)
		state.Errors = append(state.Errors, err.Error())
		if fn, ok := p.afterStage[StageGenerate]; ok {
			fn(state, err)
		}
		return nil, err
	}
	state.Hunks = hunks
	if fn, ok := p.afterStage[StageGenerate]; ok {
		fn(state, nil)
	}

	// Snapshot before verification
	var affectedFiles []string
	for _, h := range hunks {
		found := false
		for _, af := range affectedFiles {
			if af == h.FilePath {
				found = true
				break
			}
		}
		if !found {
			affectedFiles = append(affectedFiles, h.FilePath)
		}
	}

	snapshot, err := execution.CreateSnapshot(absDir, affectedFiles)
	if err != nil {
		return nil, fmt.Errorf("snapshot failed: %w", err)
	}

	// Stage: Verify — execute transaction with self-correction
	state.CurrentStage = StageVerify
	if fn, ok := p.beforeStage[StageVerify]; ok {
		fn(state)
	}

	executor := execution.NewExecutor(absDir)
	repairFunc := p.config.RepairFunc
	if repairFunc == nil {
		repairFunc = func(rCtx context.Context, prompt string, hunk execution.StructuredHunk) ([]execution.StructuredHunk, error) {
			if budgetTracker != nil && budgetTracker.IsExceeded() {
				return nil, fmt.Errorf("budget exceeded: %s", budgetTracker.ExceededMessage())
			}
			resp, err := client.Chat(rCtx, llm.ChatRequest{
				Model:       cfg.Model,
				Messages:    []llm.Message{{Role: llm.RoleSystem, Content: llm.SystemPrompt}, {Role: llm.RoleUser, Content: prompt}},
				Temperature: 0.2,
				MaxTokens:   4096,
			})
			if err != nil {
				return nil, err
			}
			if budgetTracker != nil && resp.Usage != nil {
				budgetTracker.Track(resp.Usage.PromptTokens, resp.Usage.CompletionTokens)
			}
			return parser.ParseLLMResponse(resp.Choices[0].Message.Content)
		}
	}

	summary, err := executor.ExecuteTransaction(ctx, task, absDir, hunks, execution.ExecutionContext{
		RepairFunc: repairFunc,
	})
	state.Summary = summary
	if err != nil {
		state.Errors = append(state.Errors, err.Error())
		if fn, ok := p.afterStage[StageVerify]; ok {
			fn(state, err)
		}
		return &Result{
			Status:   execution.StatusFail,
			State:    state,
			Duration: time.Since(state.StartTime),
		}, fmt.Errorf("execution failed: %w", err)
	}

	if summary.Status != execution.StatusPass {
		state.Errors = append(state.Errors, "verification failed")
		if fn, ok := p.afterStage[StageVerify]; ok {
			fn(state, fmt.Errorf("verification failed"))
		}
		return &Result{
			Status:   execution.StatusFail,
			State:    state,
			Duration: time.Since(state.StartTime),
		}, fmt.Errorf("verification failed")
	}
	if fn, ok := p.afterStage[StageVerify]; ok {
		fn(state, nil)
	}

	// Stage: Apply — automatically happens inside ExecuteTransaction commitToDisk
	state.CurrentStage = StageApply
	if fn, ok := p.beforeStage[StageApply]; ok {
		fn(state)
	}
	if fn, ok := p.afterStage[StageApply]; ok {
		fn(state, nil)
	}

	// Stage: Test — run tests, rollback on failure
	state.CurrentStage = StageTest
	if fn, ok := p.beforeStage[StageTest]; ok {
		fn(state)
	}

	testCmd := p.config.TestCommand
	if testCmd == "" {
		testCmd = execution.DetectTestCommand(absDir)
	}

	testCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	testOutput, testErr := runTestCommand(testCtx, absDir, testCmd)
	if testErr != nil {
		if restoreErr := snapshot.Restore(absDir); restoreErr != nil {
			state.Errors = append(state.Errors, fmt.Sprintf("rollback incomplete: %v", restoreErr))
		}
		state.Errors = append(state.Errors, testErr.Error())
		if fn, ok := p.afterStage[StageTest]; ok {
			fn(state, testErr)
		}
		return &Result{
			Status:     execution.StatusFail,
			State:      state,
			Duration:   time.Since(state.StartTime),
			TestOutput: testOutput,
		}, fmt.Errorf("tests failed after applying patches (rolled back)")
	}

	if fn, ok := p.afterStage[StageTest]; ok {
		fn(state, nil)
	}

	return &Result{
		Status:     execution.StatusPass,
		State:      state,
		Duration:   time.Since(state.StartTime),
		TestOutput: testOutput,
	}, nil
}

type testRunResult struct {
	output string
	err    error
}

func runTestCommand(ctx context.Context, dir string, command string) (string, error) {
	parts := execution.ParseCommand(command)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty test command")
	}

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	output := string(out)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return output, fmt.Errorf("test timed out after 120s")
		}
		return output, fmt.Errorf("test failed: %v\nOutput:\n%s", err, output)
	}

	return output, nil
}

func (p *Pipeline) Config() Config {
	return p.config
}

func (s *State) Elapsed() time.Duration {
	return time.Since(s.StartTime)
}

func (s *State) HasErrors() bool {
	return len(s.Errors) > 0
}

func (s *State) LastError() string {
	if len(s.Errors) == 0 {
		return ""
	}
	return s.Errors[len(s.Errors)-1]
}

func (s *State) WithContext(projectRoot string) *State {
	s.ProjectRoot = projectRoot
	return s
}

func AllStages() []Stage {
	return []Stage{StagePlan, StageCritique, StageGenerate, StageVerify, StageApply, StageTest}
}
