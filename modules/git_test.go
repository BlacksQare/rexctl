package modules

import (
	"os"
	"path/filepath"
	"rexctl/config"
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

func TestGitDirtyStatus_IgnoresDotEnvAndOverrides(t *testing.T) {
	tempDir := t.TempDir()

	// Initialize git repo
	_, err := runCmd(tempDir, "git", "init")
	if err != nil {
		t.Skip("git not available in environment, skipping")
	}
	runCmd(tempDir, "git", "config", "user.email", "test@example.com")
	runCmd(tempDir, "git", "config", "user.name", "Test User")

	// Create and commit a file
	testFile := filepath.Join(tempDir, "app.py")
	os.WriteFile(testFile, []byte("print('hello')"), 0644)
	runCmd(tempDir, "git", "add", "app.py")
	runCmd(tempDir, "git", "commit", "-m", "initial commit")

	commit, err := GetGitCommitHash(tempDir)
	if err != nil || commit == "" {
		t.Fatalf("failed to get commit hash: %v", err)
	}

	dummyKeyPath := filepath.Join(t.TempDir(), "authorized_keys")
	os.WriteFile(dummyKeyPath, []byte("dummy-key"), 0600)
	origKeyPath := config.DefaultAuthorizedKeysPath
	config.DefaultAuthorizedKeysPath = dummyKeyPath
	defer func() { config.DefaultAuthorizedKeysPath = origKeyPath }()

	// 1. Create .env file
	if err := WriteEnvFile(tempDir); err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}

	// Should still be clean!
	dirty, err := IsGitDirty(tempDir)
	if err != nil {
		t.Fatalf("IsGitDirty failed: %v", err)
	}
	if dirty {
		t.Errorf("expected repo with only .env to be clean, got dirty=true")
	}

	tag := BuildImageTag("myws", "app", commit, dirty)
	if tag != "rexctl/myws/app:"+commit {
		t.Errorf("expected clean image tag 'rexctl/myws/app:%s', got '%s'", commit, tag)
	}

	// 2. Create docker-compose.override.yml
	overridePath := filepath.Join(tempDir, "docker-compose.override.yml")
	os.WriteFile(overridePath, []byte("services:\n  app:\n    image: custom\n"), 0644)

	// Should still be clean!
	dirty, err = IsGitDirty(tempDir)
	if err != nil {
		t.Fatalf("IsGitDirty failed: %v", err)
	}
	if dirty {
		t.Errorf("expected repo with .env and docker-compose.override.yml to be clean, got dirty=true")
	}

	tag = BuildImageTag("myws", "app", commit, dirty)
	if tag != "rexctl/myws/app:"+commit {
		t.Errorf("expected clean image tag 'rexctl/myws/app:%s', got '%s'", commit, tag)
	}

	// 3. Modifying actual code makes it dirty
	os.WriteFile(testFile, []byte("print('changed')"), 0644)
	dirty, err = IsGitDirty(tempDir)
	if err != nil {
		t.Fatalf("IsGitDirty failed: %v", err)
	}
	if !dirty {
		t.Errorf("expected repo with modified app.py to be dirty")
	}

	dirtyTag := BuildImageTag("myws", "app", commit, dirty)
	if dirtyTag != "rexctl/myws/app:"+commit+"-dirty" {
		t.Errorf("expected dirty image tag 'rexctl/myws/app:%s-dirty', got '%s'", commit, dirtyTag)
	}
}

func TestGitDirtyStatus_OtherEnvFilesMakeDirty(t *testing.T) {
	tempDir := t.TempDir()

	_, err := runCmd(tempDir, "git", "init")
	if err != nil {
		t.Skip("git not available in environment, skipping")
	}
	runCmd(tempDir, "git", "config", "user.email", "test@example.com")
	runCmd(tempDir, "git", "config", "user.name", "Test User")

	testFile := filepath.Join(tempDir, "app.py")
	os.WriteFile(testFile, []byte("print('hello')"), 0644)
	runCmd(tempDir, "git", "add", "app.py")
	runCmd(tempDir, "git", "commit", "-m", "initial commit")

	// Create a non-.env file like custom.env or test.env
	otherEnv := filepath.Join(tempDir, "custom.env")
	os.WriteFile(otherEnv, []byte("FOO=BAR\n"), 0644)

	dirty, err := IsGitDirty(tempDir)
	if err != nil {
		t.Fatalf("IsGitDirty failed: %v", err)
	}
	if !dirty {
		t.Errorf("expected custom.env to mark repository as dirty")
	}
}

func TestGitDirtyStatus_IgnoresInitScript(t *testing.T) {
	tempDir := t.TempDir()

	_, err := runCmd(tempDir, "git", "init")
	if err != nil {
		t.Skip("git not available in environment, skipping")
	}
	runCmd(tempDir, "git", "config", "user.email", "test@example.com")
	runCmd(tempDir, "git", "config", "user.name", "Test User")

	testFile := filepath.Join(tempDir, "app.py")
	os.WriteFile(testFile, []byte("print('hello')"), 0644)
	runCmd(tempDir, "git", "add", "app.py")
	runCmd(tempDir, "git", "commit", "-m", "initial commit")

	// Create rexctl_init.sh file
	initScript := filepath.Join(tempDir, "rexctl_init.sh")
	os.WriteFile(initScript, []byte("#!/bin/bash\necho init\n"), 0755)

	dirty, err := IsGitDirty(tempDir)
	if err != nil {
		t.Fatalf("IsGitDirty failed: %v", err)
	}
	if dirty {
		t.Errorf("expected rexctl_init.sh to be ignored and repo to remain clean")
	}
}


