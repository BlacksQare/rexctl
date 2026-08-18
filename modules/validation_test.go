package modules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseManifest_Valid(t *testing.T) {
	dir := t.TempDir()
	content := `
kind: RexctlWorkspace
spec:
  containers:
    - name: api
      type: compose
      remote: https://github.com/org/api
      revision: v1.0.0
`
	os.WriteFile(filepath.Join(dir, "rex.yaml"), []byte(content), 0644)

	m, err := ParseManifest(dir)
	if err != nil {
		t.Fatalf("expected valid manifest, got: %v", err)
	}
	if m.Kind != "RexctlWorkspace" {
		t.Errorf("expected kind 'RexctlWorkspace', got '%s'", m.Kind)
	}
	if len(m.Spec.Containers) != 1 {
		t.Fatalf("expected 1 container, got %d", len(m.Spec.Containers))
	}
}

func TestParseManifest_MissingFile(t *testing.T) {
	dir := t.TempDir()
	_, err := ParseManifest(dir)
	if err == nil {
		t.Fatal("expected error for missing rex.yaml, got nil")
	}
}

func TestParseManifest_InvalidYaml(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "rex.yaml"), []byte("invalid: yaml: [unclosed"), 0644)

	_, err := ParseManifest(dir)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestValidateStackDir_Valid(t *testing.T) {
	dir := t.TempDir()
	content := `
kind: RexctlWorkspace
spec:
  containers:
    - name: svc1
      type: compose
      remote: git@github.com:org/svc1.git
      revision: main
    - name: redis
      type: image
      remote: redis:alpine
`
	os.WriteFile(filepath.Join(dir, "rex.yaml"), []byte(content), 0644)

	err := ValidateStackDir(dir)
	if err != nil {
		t.Fatalf("expected valid stack, got error: %v", err)
	}
}

func TestValidateStackDir_InvalidKind(t *testing.T) {
	dir := t.TempDir()
	content := `
kind: OtherKind
spec:
  containers:
    - name: redis
      type: image
      remote: redis:alpine
`
	os.WriteFile(filepath.Join(dir, "rex.yaml"), []byte(content), 0644)

	err := ValidateStackDir(dir)
	if err == nil || !strings.Contains(err.Error(), "RexctlWorkspace") {
		t.Fatalf("expected kind error, got: %v", err)
	}
}

func TestValidateStackDir_EmptyContainers(t *testing.T) {
	dir := t.TempDir()
	content := `
kind: RexctlWorkspace
spec:
  containers: []
`
	os.WriteFile(filepath.Join(dir, "rex.yaml"), []byte(content), 0644)

	err := ValidateStackDir(dir)
	if err == nil || !strings.Contains(err.Error(), "empty or missing") {
		t.Fatalf("expected empty containers error, got: %v", err)
	}
}

func TestValidateStackDir_MissingRequiredFields(t *testing.T) {
	dir := t.TempDir()
	content := `
kind: RexctlWorkspace
spec:
  containers:
    - name: ""
      type: image
      remote: redis:alpine
`
	os.WriteFile(filepath.Join(dir, "rex.yaml"), []byte(content), 0644)

	err := ValidateStackDir(dir)
	if err == nil || !strings.Contains(err.Error(), "missing required fields") {
		t.Fatalf("expected missing required fields error, got: %v", err)
	}
}

func TestValidateStackDir_UnknownType(t *testing.T) {
	dir := t.TempDir()
	content := `
kind: RexctlWorkspace
spec:
  containers:
    - name: app
      type: kubernetes
      remote: app:v1
`
	os.WriteFile(filepath.Join(dir, "rex.yaml"), []byte(content), 0644)

	err := ValidateStackDir(dir)
	if err == nil || !strings.Contains(err.Error(), "unknown type") {
		t.Fatalf("expected unknown type error, got: %v", err)
	}
}

func TestValidateStackDir_ComposeMissingRevision(t *testing.T) {
	dir := t.TempDir()
	content := `
kind: RexctlWorkspace
spec:
  containers:
    - name: app
      type: compose
      remote: git@github.com:org/app.git
`
	os.WriteFile(filepath.Join(dir, "rex.yaml"), []byte(content), 0644)

	err := ValidateStackDir(dir)
	if err == nil || !strings.Contains(err.Error(), "missing required field 'revision'") {
		t.Fatalf("expected missing revision error, got: %v", err)
	}
}

func TestValidateStackDir_DuplicateContainerNames(t *testing.T) {
	dir := t.TempDir()
	content := `
kind: RexctlWorkspace
spec:
  containers:
    - name: redis
      type: image
      remote: redis:alpine
    - name: redis
      type: image
      remote: redis:7
`
	os.WriteFile(filepath.Join(dir, "rex.yaml"), []byte(content), 0644)

	err := ValidateStackDir(dir)
	if err == nil || !strings.Contains(err.Error(), "duplicated") {
		t.Fatalf("expected duplicate name error, got: %v", err)
	}
}

func TestGetTargetStack_WithArg(t *testing.T) {
	stack, err := GetTargetStack("my-custom-stack")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if stack != "my-custom-stack" {
		t.Errorf("expected 'my-custom-stack', got '%s'", stack)
	}
}

func TestGetTargetStack_NoArgNoRexYaml(t *testing.T) {
	tempDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(origWd)

	_, err := GetTargetStack("")
	if err == nil {
		t.Fatal("expected error when no arg passed and no rex.yaml in cwd, got nil")
	}
}

func TestGetTargetStack_NoArgWithRexYamlInCwd(t *testing.T) {
	tempDir := t.TempDir()
	os.WriteFile(filepath.Join(tempDir, "rex.yaml"), []byte("kind: RexctlWorkspace"), 0644)
	origWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(origWd)

	stack, err := GetTargetStack("")
	if err != nil {
		t.Fatalf("expected success with rex.yaml in cwd, got: %v", err)
	}
	expected := filepath.Base(tempDir)
	if stack != expected {
		t.Errorf("expected '%s', got '%s'", expected, stack)
	}
}

func TestValidateStack_PanicsOnInvalid(t *testing.T) {
	origDie := Die
	died := false
	Die = func(format string, a ...any) {
		died = true
		panic("died")
	}
	defer func() {
		Die = origDie
		if r := recover(); r == nil {
			t.Fatal("expected ValidateStack to call Die on nonexistent dir")
		}
		if !died {
			t.Error("expected Die to be called")
		}
	}()

	ValidateStack("/nonexistent/workspace/dir/xyz")
}

func TestResolveWorkspace_DirectDirectory(t *testing.T) {
	tempDir := t.TempDir()
	os.WriteFile(filepath.Join(tempDir, "rex.yaml"), []byte("kind: RexctlWorkspace"), 0644)

	name, dir, err := ResolveWorkspace(tempDir)
	if err != nil {
		t.Fatalf("expected ResolveWorkspace on direct dir to succeed, got: %v", err)
	}
	if dir != tempDir {
		t.Errorf("expected dir '%s', got '%s'", tempDir, dir)
	}
	if name != filepath.Base(tempDir) {
		t.Errorf("expected name '%s', got '%s'", filepath.Base(tempDir), name)
	}
}


