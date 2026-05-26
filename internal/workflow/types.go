package workflow

import (
	"time"

	"github.com/kode/kode/internal/execution"
	"github.com/kode/kode/internal/llm"
	"github.com/kode/kode/internal/router"
)

type Stage string

const (
	StagePlan     Stage = "plan"
	StageCritique Stage = "critique"
	StageGenerate Stage = "generate"
	StageVerify   Stage = "verify"
	StageApply    Stage = "apply"
	StageTest     Stage = "test"
)

type State struct {
	CurrentStage Stage
	TaskID       string
	ProjectRoot  string
	Task         string
	Hunks        []execution.StructuredHunk
	Summary      *execution.ExecutionSummary
	Errors       []string
	StartTime    time.Time
}

type Config struct {
	LLMConfig          *llm.Config
	MaxRetries         int
	TestCommand        string
	ModelOverride      string
	ContextFile        string
	RepairFunc         execution.RepairFunc
	TokenBudget        *llm.TokenBudget
	EnableContextIndex bool
	RouterConfig       *router.RouteConfig
}

type Result struct {
	Status     execution.Status
	State      *State
	Duration   time.Duration
	TestOutput string
}

type Pipeline struct {
	config      Config
	beforeStage map[Stage]func(*State)
	afterStage  map[Stage]func(*State, error)
}

type StageHook func(state *State) error
