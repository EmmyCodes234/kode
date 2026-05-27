package workflow

import (
	"testing"
)

func TestNewPipeline(t *testing.T) {
	cfg := Config{MaxRetries: 5}
	p := NewPipeline(cfg)
	if p.config.MaxRetries != 5 {
		t.Errorf("Expected MaxRetries 5, got %d", p.config.MaxRetries)
	}
}

func TestPipelineHooks(t *testing.T) {
	p := NewPipeline(Config{})
	
	beforeCalled := false
	p.BeforeStage(StagePlan, func(s *State) {
		beforeCalled = true
	})

	afterCalled := false
	p.AfterStage(StagePlan, func(s *State, err error) {
		afterCalled = true
	})

	// Simulate calling the hooks directly since Run would require real LLM/file setup
	if fn, ok := p.beforeStage[StagePlan]; ok {
		fn(&State{})
	}
	if fn, ok := p.afterStage[StagePlan]; ok {
		fn(&State{}, nil)
	}

	if !beforeCalled {
		t.Error("BeforeStage hook not called")
	}
	if !afterCalled {
		t.Error("AfterStage hook not called")
	}
}
