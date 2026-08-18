package modules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListWorkspaces_Multiple(t *testing.T) {
	tempDir := t.TempDir()

	// Create ws1 with rex.yaml
	ws1 := filepath.Join(tempDir, "ws1")
	os.MkdirAll(ws1, 0755)
	os.WriteFile(filepath.Join(ws1, "rex.yaml"), []byte("kind: RexctlWorkspace"), 0644)

	// Create ws2 with rex.yaml
	ws2 := filepath.Join(tempDir, "ws2")
	os.MkdirAll(ws2, 0755)
	os.WriteFile(filepath.Join(ws2, "rex.yaml"), []byte("kind: RexctlWorkspace"), 0644)

	// Create regular folder without rex.yaml (should be ignored)
	ignored := filepath.Join(tempDir, "ignored_dir")
	os.MkdirAll(ignored, 0755)

	// Create a regular file in workspaces dir
	os.WriteFile(filepath.Join(tempDir, "somefile.txt"), []byte("text"), 0644)

	list, err := ListWorkspaces(tempDir)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}

	if len(list) != 2 {
		t.Fatalf("expected 2 workspaces, got %d: %v", len(list), list)
	}

	foundWs1, foundWs2 := false, false
	for _, name := range list {
		if name == "ws1" {
			foundWs1 = true
		}
		if name == "ws2" {
			foundWs2 = true
		}
	}

	if !foundWs1 || !foundWs2 {
		t.Errorf("missing expected workspaces in list: %v", list)
	}
}

func TestListWorkspaces_EmptyDir(t *testing.T) {
	tempDir := t.TempDir()
	list, err := ListWorkspaces(tempDir)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected 0 workspaces, got %d", len(list))
	}
}

func TestListWorkspaces_NonExistentDir(t *testing.T) {
	_, err := ListWorkspaces("/non/existent/path/xyz-123")
	if err == nil {
		t.Fatal("expected error for non-existent directory, got nil")
	}
}
