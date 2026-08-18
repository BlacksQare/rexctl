package modules

import (
	"testing"
)

func TestParseActiveStackFromOutput_SingleRunning(t *testing.T) {
	out := "my-robot-stack\n"
	stack, found := ParseActiveStackFromOutput(out)
	if !found {
		t.Fatal("expected found to be true")
	}
	if stack != "my-robot-stack" {
		t.Errorf("expected 'my-robot-stack', got '%s'", stack)
	}
}

func TestParseActiveStackFromOutput_MultipleContainersSameStack(t *testing.T) {
	out := "my-robot-stack\nmy-robot-stack\nmy-robot-stack\n"
	stack, found := ParseActiveStackFromOutput(out)
	if !found {
		t.Fatal("expected found to be true")
	}
	if stack != "my-robot-stack" {
		t.Errorf("expected 'my-robot-stack', got '%s'", stack)
	}
}

func TestParseActiveStackFromOutput_Empty(t *testing.T) {
	out := "\n  \n\n"
	stack, found := ParseActiveStackFromOutput(out)
	if found {
		t.Errorf("expected found to be false, got true with stack '%s'", stack)
	}
	if stack != "" {
		t.Errorf("expected empty stack string, got '%s'", stack)
	}
}

func TestParseActiveStackFromOutput_WithRexctlLabel(t *testing.T) {
	out := "my-workspace|compose-project-name\n"
	stack, found := ParseActiveStackFromOutput(out)
	if !found {
		t.Fatal("expected found to be true")
	}
	if stack != "my-workspace" {
		t.Errorf("expected 'my-workspace', got '%s'", stack)
	}
}

func TestResolveRunningContainer_Fallback(t *testing.T) {
	name := ResolveRunningContainer("test-ws", "nonexistent-container")
	if name != "nonexistent-container" {
		t.Errorf("expected fallback to 'nonexistent-container', got '%s'", name)
	}
}

