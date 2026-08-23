package modules

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"rexctl/config"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func captureStdout(f func()) string {
	r, w, _ := os.Pipe()
	stdout := os.Stdout
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = stdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestCreateWorkspace_Success(t *testing.T) {
	tempDir := t.TempDir()
	origWorkspaces := config.WorkspacesDir
	config.WorkspacesDir = tempDir
	defer func() { config.WorkspacesDir = origWorkspaces }()

	manifest := []byte("kind: RexctlWorkspace\nspec:\n  containers: []\n")
	err := CreateWorkspace("new-ws", manifest)
	if err != nil {
		t.Fatalf("expected CreateWorkspace to succeed, got: %v", err)
	}

	createdFile := filepath.Join(tempDir, "new-ws", "rex.yaml")
	data, err := os.ReadFile(createdFile)
	if err != nil {
		t.Fatalf("expected rex.yaml to exist: %v", err)
	}

	if string(data) != string(manifest) {
		t.Errorf("manifest content mismatch: got %s", string(data))
	}
}

func TestCmdPwd_Output(t *testing.T) {
	origWorkspaces := config.WorkspacesDir
	config.WorkspacesDir = "/test/workspaces"
	defer func() { config.WorkspacesDir = origWorkspaces }()

	out := captureStdout(func() {
		CmdPwd([]string{"robotics"})
	})

	expected := filepath.Join("/test/workspaces", "robotics") + "\n"
	if out != expected {
		t.Errorf("expected '%s', got '%s'", expected, out)
	}
}

func TestCmdGet_EmptyWhenNoActiveStack(t *testing.T) {
	// If getActiveStack returns "", CmdGet must output nothing (empty string)
	out := captureStdout(func() {
		CmdGet()
	})

	// When running without active docker compose container, output should be empty string
	if strings.Contains(out, "No workspace is currently running.") {
		t.Errorf("CmdGet should NOT output 'No workspace is currently running.', got: '%s'", out)
	}
}

func TestWorkspaceFullLifecycle_CreateSyncValidateInfo(t *testing.T) {
	tempDir := t.TempDir()
	origWorkspaces := config.WorkspacesDir
	config.WorkspacesDir = tempDir
	defer func() { config.WorkspacesDir = origWorkspaces }()

	stackName := "robotics-ws"
	manifestContent := `kind: RexctlWorkspace
spec:
  containers:
    - name: vision_svc
      type: compose
      remote: git@github.com:org/vision.git
      revision: main
    - name: redis_db
      type: image
      remote: redis:7
`
	// 1. Create Workspace
	err := CreateWorkspace(stackName, []byte(manifestContent))
	if err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	stackDir := filepath.Join(tempDir, stackName)
	visionDir := filepath.Join(stackDir, "vision_svc")
	os.MkdirAll(visionDir, 0755)

	// Create git repository in vision_svc
	if _, err := runCmd(visionDir, "git", "init"); err != nil {
		t.Skip("git not available in environment, skipping")
	}
	runCmd(visionDir, "git", "config", "user.email", "test@example.com")
	runCmd(visionDir, "git", "config", "user.name", "Tester")
	composeYAML := "services:\n  vision:\n    build: .\n"
	os.WriteFile(filepath.Join(visionDir, "docker-compose.yml"), []byte(composeYAML), 0644)
	runCmd(visionDir, "git", "add", "docker-compose.yml")
	runCmd(visionDir, "git", "commit", "-m", "initial")

	// 2. Validate Stack
	err = ValidateStackDir(stackDir)
	if err != nil {
		t.Fatalf("stack validation failed: %v", err)
	}

	// 3. Write Standard Override (Sync phase)
	commit, err := GetGitCommitHash(visionDir)
	if err != nil || commit == "" {
		t.Fatalf("failed to get commit hash: %v", err)
	}
	dirty, _ := IsGitDirty(visionDir)

	err = WriteStandardOverride(visionDir, stackName, "vision_svc", commit, dirty)
	if err != nil {
		t.Fatalf("failed to write standard override: %v", err)
	}

	// Verify docker-compose.override.yml was written with correct workspace-specific image name
	overrideData, err := os.ReadFile(filepath.Join(visionDir, "docker-compose.override.yml"))
	if err != nil {
		t.Fatalf("failed to read override: %v", err)
	}

	var parsed StandardOverride
	if err := yaml.Unmarshal(overrideData, &parsed); err != nil {
		t.Fatalf("failed to parse standard override: %v", err)
	}

	if parsed.Name != stackName {
		t.Errorf("expected name '%s', got '%s'", stackName, parsed.Name)
	}

	visionSvc, ok := parsed.Services["vision"]
	if !ok {
		t.Fatal("expected 'vision' service in standard override")
	}

	if visionSvc.ContainerName != stackName+"-vision" {
		t.Errorf("expected container_name '%s-vision', got '%s'", stackName, visionSvc.ContainerName)
	}

	expectedImage := "rexctl/" + stackName + "/vision:" + commit
	if visionSvc.Image != expectedImage {
		t.Errorf("expected image '%s', got '%s'", expectedImage, visionSvc.Image)
	}
	if visionSvc.Labels["rexctl.workspace"] != stackName {
		t.Errorf("expected label rexctl.workspace='%s', got '%s'", stackName, visionSvc.Labels["rexctl.workspace"])
	}

	// 4. Test Workspace Info inspection (Clean State)
	info, err := GetWorkspaceInfo(stackDir, stackName, stackName)
	if err != nil {
		t.Fatalf("failed to get workspace info: %v", err)
	}

	if !info.IsRunning {
		t.Error("expected IsRunning to be true")
	}
	if len(info.Containers) != 2 {
		t.Fatalf("expected 2 containers, got %d", len(info.Containers))
	}

	c0 := info.Containers[0]
	expectedCleanInfoImage := "rexctl/" + stackName + "/vision_svc:" + commit
	if c0.Name != "vision_svc" || c0.ImageTag != expectedCleanInfoImage || c0.State != "clean" {
		t.Errorf("unexpected container 0 info in clean state: %+v", c0)
	}

	c1 := info.Containers[1]
	expectedRawImage := "rexctl/" + stackName + "/redis_db:latest"
	if c1.Name != "redis_db" || c1.ImageTag != expectedRawImage {
		t.Errorf("unexpected container 1 info: %+v", c1)
	}

	// 5. Test Modification -> Dirty State Detection
	os.WriteFile(filepath.Join(visionDir, "modified_code.py"), []byte("print('modified')"), 0644)
	dirtyAfterMod, err := IsGitDirty(visionDir)
	if err != nil || !dirtyAfterMod {
		t.Fatalf("expected repo to be dirty after modification, got dirty=%t, err=%v", dirtyAfterMod, err)
	}

	err = WriteStandardOverride(visionDir, stackName, "vision_svc", commit, dirtyAfterMod)
	if err != nil {
		t.Fatalf("failed to update standard override: %v", err)
	}

	overrideDataDirty, err := os.ReadFile(filepath.Join(visionDir, "docker-compose.override.yml"))
	if err != nil {
		t.Fatalf("failed to read override: %v", err)
	}
	var parsedDirty StandardOverride
	yaml.Unmarshal(overrideDataDirty, &parsedDirty)

	expectedDirtyImage := "rexctl/" + stackName + "/vision:" + commit + "-dirty"
	if parsedDirty.Services["vision"].Image != expectedDirtyImage {
		t.Errorf("expected dirty image '%s', got '%s'", expectedDirtyImage, parsedDirty.Services["vision"].Image)
	}
	if parsedDirty.Services["vision"].Labels["rexctl.dirty"] != "true" {
		t.Errorf("expected dirty label 'true'")
	}



	// 5. Test Compose Lifecycle Argument Generation (Up, Start, Stop, Down)
	upArgs := BuildComposeArgs(stackName, "up", "-d")
	if strings.Join(upArgs, " ") != "compose --project-name robotics-ws up -d" {
		t.Errorf("unexpected upArgs: %v", upArgs)
	}

	startArgs := BuildComposeArgs(stackName, "start")
	if strings.Join(startArgs, " ") != "compose --project-name robotics-ws start" {
		t.Errorf("unexpected startArgs: %v", startArgs)
	}

	stopArgs := BuildComposeArgs(stackName, "stop")
	if strings.Join(stopArgs, " ") != "compose --project-name robotics-ws stop" {
		t.Errorf("unexpected stopArgs: %v", stopArgs)
	}

	downArgs := BuildComposeArgs(stackName, "down")
	if strings.Join(downArgs, " ") != "compose --project-name robotics-ws down" {
		t.Errorf("unexpected downArgs: %v", downArgs)
	}
}

func TestWorkspaceLifecycle_MultiWorkspaceImageIsolation(t *testing.T) {
	ws1 := "workspace-dev"
	ws2 := "workspace-prod"
	serviceName := "core_engine"
	commit := "abc5678"

	services := []ServiceBuildInfo{{Name: serviceName, HasBuild: true}}
	override1, err := GenerateStandardOverride(ws1, services, commit, false)
	if err != nil {
		t.Fatalf("failed to generate override for ws1: %v", err)
	}

	override2, err := GenerateStandardOverride(ws2, services, commit, true)
	if err != nil {
		t.Fatalf("failed to generate override for ws2: %v", err)
	}


	var parsed1, parsed2 StandardOverride
	yaml.Unmarshal([]byte(override1), &parsed1)
	yaml.Unmarshal([]byte(override2), &parsed2)

	img1 := parsed1.Services[serviceName].Image
	img2 := parsed2.Services[serviceName].Image

	if img1 != "rexctl/workspace-dev/core_engine:abc5678" {
		t.Errorf("unexpected img1: %s", img1)
	}
	if img2 != "rexctl/workspace-prod/core_engine:abc5678-dirty" {
		t.Errorf("unexpected img2: %s", img2)
	}
	if img1 == img2 {
		t.Errorf("expected images for ws1 and ws2 to be strictly isolated")
	}
}

func TestCmdDestroy_Validation(t *testing.T) {
	origDie := Die
	died := false
	Die = func(format string, a ...any) {
		died = true
		panic("died")
	}
	defer func() {
		Die = origDie
	}()

	// No args -> Die
	func() {
		defer func() { recover() }()
		CmdDestroy([]string{})
	}()
	if !died {
		t.Error("expected Die when no args passed to CmdDestroy")
	}

	// Non-existent stack -> Die
	died = false
	func() {
		defer func() { recover() }()
		CmdDestroy([]string{"nonexistent-destroy-ws"})
	}()
	if !died {
		t.Error("expected Die when nonexistent stack passed to CmdDestroy")
	}
}

func TestCmdStatus_Output(t *testing.T) {
	out := captureStdout(func() {
		CmdStatus()
	})
	if out == "" {
		t.Error("expected CmdStatus to produce output")
	}
}

func TestCmdList_Output(t *testing.T) {
	tempDir := t.TempDir()
	origWorkspaces := config.WorkspacesDir
	config.WorkspacesDir = tempDir
	defer func() { config.WorkspacesDir = origWorkspaces }()

	wsDir := filepath.Join(tempDir, "list-ws")
	os.MkdirAll(wsDir, 0755)
	os.WriteFile(filepath.Join(wsDir, "rex.yaml"), []byte("kind: RexctlWorkspace"), 0644)

	out := captureStdout(func() {
		CmdList()
	})
	if !strings.Contains(out, "list-ws") {
		t.Errorf("expected CmdList to include 'list-ws', got '%s'", out)
	}
}

func TestCmdSwitch_Validation(t *testing.T) {
	origDie := Die
	died := false
	Die = func(format string, a ...any) {
		died = true
		panic("died")
	}
	defer func() {
		Die = origDie
	}()

	// No args -> Die
	func() {
		defer func() { recover() }()
		CmdSwitch([]string{})
	}()
	if !died {
		t.Error("expected Die when no args passed to CmdSwitch")
	}
}

