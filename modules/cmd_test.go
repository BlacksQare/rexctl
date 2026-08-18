package modules

import (
	"path/filepath"
	"testing"
)


func TestRunCmd_Success(t *testing.T) {
	out, err := runCmd(".", "echo", "hello world")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if out != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", out)
	}
}

func TestRunCmd_Failure(t *testing.T) {
	_, err := runCmd(".", "false")
	if err == nil {
		t.Fatal("expected error from false command, got nil")
	}
}

func TestRunCmd_WithDir(t *testing.T) {
	tempDir := t.TempDir()
	out, err := runCmd(tempDir, "pwd")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	// On Linux / macOS resolving symlinks
	realTemp, _ := filepath.EvalSymlinks(tempDir)
	realOut, _ := filepath.EvalSymlinks(out)
	if realTemp != realOut {
		t.Errorf("expected '%s', got '%s'", realTemp, realOut)
	}
}

func TestRunCmdStream_Success(t *testing.T) {
	err := runCmdStream(".", "echo", "streaming")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
}

func TestRunCmdStream_Failure(t *testing.T) {
	err := runCmdStream(".", "false")
	if err == nil {
		t.Fatal("expected error from failing stream command, got nil")
	}
}
