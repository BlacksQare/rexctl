package modules

import (
	"os"
	"path/filepath"
	"rexctl/config"
	"testing"
)

func TestPrepareEnv_ExecutesAllInitScripts(t *testing.T) {
	stackDir := t.TempDir()

	manifest := `
kind: RexctlWorkspace
spec:
  containers:
    - name: repo1
      type: compose
      remote: git@github.com:org/repo1.git
      revision: main
    - name: repo2
      type: compose
      remote: git@github.com:org/repo2.git
      revision: main
`
	os.WriteFile(filepath.Join(stackDir, "rex.yaml"), []byte(manifest), 0644)

	repo1Dir := filepath.Join(stackDir, "repo1")
	repo2Dir := filepath.Join(stackDir, "repo2")
	os.MkdirAll(repo1Dir, 0755)
	os.MkdirAll(repo2Dir, 0755)

	// Create rexctl_init.sh in repo1 (creates marker1.txt)
	init1 := filepath.Join(repo1Dir, config.DefaultInitScriptName)
	os.WriteFile(init1, []byte("#!/bin/bash\necho 'done1' > marker1.txt\n"), 0755)

	// Create rexctl_init.sh in repo2 (creates marker2.txt)
	init2 := filepath.Join(repo2Dir, config.DefaultInitScriptName)
	os.WriteFile(init2, []byte("#!/bin/bash\necho 'done2' > marker2.txt\n"), 0755)

	err := RunPrepareEnv(stackDir)
	if err != nil {
		t.Fatalf("expected prepare-env to succeed, got: %v", err)
	}

	// Verify marker1 was created in repo1Dir
	if _, err := os.Stat(filepath.Join(repo1Dir, "marker1.txt")); os.IsNotExist(err) {
		t.Errorf("expected marker1.txt to be created by repo1 init script")
	}

	// Verify marker2 was created in repo2Dir
	if _, err := os.Stat(filepath.Join(repo2Dir, "marker2.txt")); os.IsNotExist(err) {
		t.Errorf("expected marker2.txt to be created by repo2 init script")
	}
}

func TestPrepareEnv_SkipsReposWithoutScript(t *testing.T) {
	stackDir := t.TempDir()

	manifest := `
kind: RexctlWorkspace
spec:
  containers:
    - name: repo_no_script
      type: compose
      remote: git@github.com:org/repo.git
      revision: main
`
	os.WriteFile(filepath.Join(stackDir, "rex.yaml"), []byte(manifest), 0644)

	repoDir := filepath.Join(stackDir, "repo_no_script")
	os.MkdirAll(repoDir, 0755)

	err := RunPrepareEnv(stackDir)
	if err != nil {
		t.Fatalf("expected prepare-env to succeed without init script, got: %v", err)
	}
}

func TestPrepareEnv_ScriptFailureHandled(t *testing.T) {
	stackDir := t.TempDir()

	manifest := `
kind: RexctlWorkspace
spec:
  containers:
    - name: failing_repo
      type: compose
      remote: git@github.com:org/failing.git
      revision: main
`
	os.WriteFile(filepath.Join(stackDir, "rex.yaml"), []byte(manifest), 0644)

	repoDir := filepath.Join(stackDir, "failing_repo")
	os.MkdirAll(repoDir, 0755)

	initScript := filepath.Join(repoDir, config.DefaultInitScriptName)
	os.WriteFile(initScript, []byte("#!/bin/bash\nexit 1\n"), 0755)

	err := RunPrepareEnv(stackDir)
	if err == nil {
		t.Fatal("expected error when init script exits with non-zero code, got nil")
	}
}
