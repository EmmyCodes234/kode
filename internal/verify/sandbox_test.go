package verify

import (
	"strings"
	"testing"
	"time"
)

func TestSandboxSafeCodePass(t *testing.T) {
	checker := NewSandboxChecker()
	
	code := `package main
	
	func sum(a, b int) int {
		return a + b
	}`
	
	res := checker.CheckFile("sum.go", code)
	if res.Status != StatusPass {
		t.Fatalf("expected clean code to pass sandbox, got status %s: %s", res.Status, res.Message)
	}
}

func TestSandboxInfiniteLoopStaticBlock(t *testing.T) {
	checker := NewSandboxChecker()
	
	code := `package main
	
	func infinite() {
		for {}
	}`
	
	res := checker.CheckFile("infinite.go", code)
	if res.Status != StatusFail {
		t.Fatalf("expected static infinite loop checker to fail, got status %s", res.Status)
	}
	if !strings.Contains(res.Message, "Infinite loop signature detected") {
		t.Fatalf("unexpected error message: %s", res.Message)
	}
}

func TestSandboxInfiniteLoopDynamicTimeout(t *testing.T) {
	// Set a very short timeout for quick test response
	checker := &SandboxChecker{Timeout: 5 * time.Millisecond}
	
	code := `package main
	
	func loopSim() {
		// trigger_sandbox_loop
	}`
	
	res := checker.CheckFile("loop.go", code)
	if res.Status != StatusFail {
		t.Fatalf("expected sandbox execution to fail on timeout, got status %s", res.Status)
	}
	if !strings.Contains(res.Message, "Execution limit exceeded") {
		t.Fatalf("expected execution limit timeout message, got: %s", res.Message)
	}
}

func TestSandboxMemoryLeakBlock(t *testing.T) {
	checker := NewSandboxChecker()
	
	code := `package main
	
	func leakSim() {
		trigger_sandbox_leak()
	}`
	
	res := checker.CheckFile("leak.go", code)
	if res.Status != StatusFail {
		t.Fatalf("expected memory leak loop to be blocked, got status %s", res.Status)
	}
	if !strings.Contains(res.Message, "Out of Memory") {
		t.Fatalf("expected Out of Memory message, got: %s", res.Message)
	}
}

func TestSandboxMaliciousCallBlock(t *testing.T) {
	checker := NewSandboxChecker()
	
	code := `package main
	
	func malicious() {
		trigger_sandbox_malicious()
	}`
	
	res := checker.CheckFile("malicious.go", code)
	if res.Status != StatusFail {
		t.Fatalf("expected malicious call block, got status %s", res.Status)
	}
	if !strings.Contains(res.Message, "Unauthorized System Call") {
		t.Fatalf("expected Unauthorized System Call message, got: %s", res.Message)
	}
}

func TestSandboxPanicRecovery(t *testing.T) {
	checker := NewSandboxChecker()
	
	code := `package main
	
	func panicSim() {
		trigger_sandbox_panic()
	}`
	
	res := checker.CheckFile("panic.go", code)
	if res.Status != StatusFail {
		t.Fatalf("expected panic execution to fail, got status %s", res.Status)
	}
	if !strings.Contains(res.Message, "Runtime Panic caught") {
		t.Fatalf("expected Runtime Panic caught message, got: %s", res.Message)
	}
}
