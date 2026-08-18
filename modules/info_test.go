package modules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetWorkspaceInfo_ValidWorkspace(t *testing.T) {
	dir := t.TempDir()

	manifest := `
kind: RexctlWorkspace
spec:
  containers:
    - name: raptor_ws
      type: compose
      remote: https://github.com/example/raptor_ws
      revision: master
    - name: nginx
      type: image
      remote: nginx:latest
`
	os.WriteFile(filepath.Join(dir, "rex.yaml"), []byte(manifest), 0644)

	// Create raptor_ws git repo
	raptorDir := filepath.Join(dir, "raptor_ws")
	os.MkdirAll(raptorDir, 0755)
	if _, err := runCmd(raptorDir, "git", "init"); err != nil {
		t.Skip("git not available in environment, skipping")
	}
	runCmd(raptorDir, "git", "config", "user.email", "test@example.com")
	runCmd(raptorDir, "git", "config", "user.name", "Test")
	os.WriteFile(filepath.Join(raptorDir, "README.md"), []byte("raptor"), 0644)
	runCmd(raptorDir, "git", "add", "README.md")
	runCmd(raptorDir, "git", "commit", "-m", "init")

	info, err := GetWorkspaceInfo(dir, "test-ws", "test-ws")
	if err != nil {
		t.Fatalf("expected GetWorkspaceInfo to succeed, got: %v", err)
	}

	if info.StackName != "test-ws" {
		t.Errorf("expected StackName 'test-ws', got '%s'", info.StackName)
	}
	if !info.IsRunning {
		t.Error("expected IsRunning to be true when activeStack == stackName")
	}

	if len(info.Containers) != 2 {
		t.Fatalf("expected 2 containers, got %d", len(info.Containers))
	}

	c0 := info.Containers[0]
	if c0.Name != "raptor_ws" || c0.Type != "compose" || c0.RequestedRevision != "master" {
		t.Errorf("unexpected c0: %+v", c0)
	}
	if c0.CurrentRevision == "MISSING" || c0.CurrentRevision == "" {
		t.Errorf("expected valid current revision, got: %s", c0.CurrentRevision)
	}
	if c0.State != "clean" {
		t.Errorf("expected clean state, got: %s", c0.State)
	}

	c1 := info.Containers[1]
	if c1.Name != "nginx" || c1.Type != "image" || c1.State != "remote image" {
		t.Errorf("unexpected c1: %+v", c1)
	}
}

func TestGetWorkspaceInfo_NotRunning(t *testing.T) {
	dir := t.TempDir()
	manifest := `
kind: RexctlWorkspace
spec:
  containers:
    - name: redis
      type: image
      remote: redis:alpine
`
	os.WriteFile(filepath.Join(dir, "rex.yaml"), []byte(manifest), 0644)

	info, err := GetWorkspaceInfo(dir, "inactive-ws", "other-ws")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if info.IsRunning {
		t.Error("expected IsRunning to be false when active stack is different")
	}
}

func TestGetWorkspaceInfo_MissingManifest(t *testing.T) {
	dir := t.TempDir()
	_, err := GetWorkspaceInfo(dir, "empty-ws", "")
	if err == nil {
		t.Fatal("expected error when rex.yaml is missing, got nil")
	}
}
