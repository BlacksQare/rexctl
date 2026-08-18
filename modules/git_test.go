package modules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildImageTag_CleanCommit(t *testing.T) {
	tag := BuildImageTag("ws1", "backend", "a1b2c3d", false)
	expected := "rexctl/ws1/backend:a1b2c3d"
	if tag != expected {
		t.Errorf("expected '%s', got '%s'", expected, tag)
	}
}

func TestBuildImageTag_DirtyCommit(t *testing.T) {
	tag := BuildImageTag("ws1", "backend", "a1b2c3d", true)
	expected := "rexctl/ws1/backend:a1b2c3d-dirty"
	if tag != expected {
		t.Errorf("expected '%s', got '%s'", expected, tag)
	}
}

func TestBuildImageTag_EmptyCommit(t *testing.T) {
	tag := BuildImageTag("ws1", "nginx", "", false)
	expected := "rexctl/ws1/nginx:latest"
	if tag != expected {
		t.Errorf("expected '%s', got '%s'", expected, tag)
	}
}

func TestImagePerWorkspaceRecognition(t *testing.T) {
	tagWsA := BuildImageTag("alpha", "robot_core", "commit1", false)
	tagWsB := BuildImageTag("beta", "robot_core", "commit1", false)

	if tagWsA == tagWsB {
		t.Errorf("expected image tags for different workspaces to be distinct, got '%s' for both", tagWsA)
	}
	if tagWsA != "rexctl/alpha/robot_core:commit1" {
		t.Errorf("expected 'rexctl/alpha/robot_core:commit1', got '%s'", tagWsA)
	}
	if tagWsB != "rexctl/beta/robot_core:commit1" {
		t.Errorf("expected 'rexctl/beta/robot_core:commit1', got '%s'", tagWsB)
	}
}

func TestGitCommitAndDirtyStatus_RealRepo(t *testing.T) {
	tempDir := t.TempDir()

	// Initialize git repo
	_, err := runCmd(tempDir, "git", "init")
	if err != nil {
		t.Skip("git not available in environment, skipping")
	}
	runCmd(tempDir, "git", "config", "user.email", "test@example.com")
	runCmd(tempDir, "git", "config", "user.name", "Test User")

	// Create and commit a file
	testFile := filepath.Join(tempDir, "file.txt")
	os.WriteFile(testFile, []byte("initial"), 0644)
	runCmd(tempDir, "git", "add", "file.txt")
	runCmd(tempDir, "git", "commit", "-m", "initial commit")

	// Check clean status
	commit, err := GetGitCommitHash(tempDir)
	if err != nil {
		t.Fatalf("failed to get commit hash: %v", err)
	}
	if len(commit) == 0 {
		t.Fatal("expected non-empty commit hash")
	}

	dirty, err := IsGitDirty(tempDir)
	if err != nil {
		t.Fatalf("failed to check dirty state: %v", err)
	}
	if dirty {
		t.Error("expected repository to be clean")
	}

	// Make changes (dirty)
	os.WriteFile(testFile, []byte("modified"), 0644)
	dirty, err = IsGitDirty(tempDir)
	if err != nil {
		t.Fatalf("failed to check dirty state after modification: %v", err)
	}
	if !dirty {
		t.Error("expected repository to be dirty after modification")
	}

	// Tag should now include -dirty
	tag := BuildImageTag("testws", "service", commit, dirty)
	if !strings.HasSuffix(tag, "-dirty") {
		t.Errorf("expected tag '%s' to have -dirty suffix", tag)
	}
	if !strings.HasPrefix(tag, "rexctl/testws/service:") {
		t.Errorf("expected tag '%s' to have prefix 'rexctl/testws/service:'", tag)
	}
}
