package daemon

import (
	"context"
	"testing"
	"time"
)

func TestDaemonInitialization(t *testing.T) {
	cfg := DaemonConfig{
		RepoDir:      "../..",
		PollInterval: time.Second,
	}
	d := NewDaemon(cfg)
	if d == nil {
		t.Fatal("NewDaemon returned nil")
	}
	if d.interval != time.Second {
		t.Errorf("Expected interval 1s, got %v", d.interval)
	}
}

func TestDaemonLifecycle(t *testing.T) {
	cfg := DaemonConfig{
		RepoDir:      "../..",
		PollInterval: time.Millisecond * 10,
	}
	d := NewDaemon(cfg)
	
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond * 50)
	defer cancel()

	err := d.Run(ctx)
	if err != nil && err != context.DeadlineExceeded && err != context.Canceled {
		t.Errorf("Unexpected error from daemon Run: %v", err)
	}
}
