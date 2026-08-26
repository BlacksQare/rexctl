package modules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteEnvFile_NewFile(t *testing.T) {
	dir := t.TempDir()
	err := WriteEnvFile(dir)
	if err != nil {
		t.Fatalf("expected WriteEnvFile to succeed, got: %v", err)
	}

	envPath := filepath.Join(dir, ".env")
	data, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("failed to read .env: %v", err)
	}

	content := string(data)
	expected := "REX_CONTAINER_AUTHORIZED_KEYS=~/.ssh/authorized_keys\n"
	if content != expected {
		t.Errorf("expected '%s', got '%s'", expected, content)
	}
}

func TestWriteEnvFile_ExistingFile_Appends(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	initialContent := "PORT=8080\nDEBUG=true\n"
	if err := os.WriteFile(envPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to create initial .env: %v", err)
	}

	err := WriteEnvFile(dir)
	if err != nil {
		t.Fatalf("expected WriteEnvFile to succeed, got: %v", err)
	}

	data, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("failed to read .env: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "PORT=8080") {
		t.Errorf("expected PORT=8080 to be preserved")
	}
	if !strings.Contains(content, "DEBUG=true") {
		t.Errorf("expected DEBUG=true to be preserved")
	}
	if !strings.Contains(content, "REX_CONTAINER_AUTHORIZED_KEYS=~/.ssh/authorized_keys") {
		t.Errorf("expected REX_CONTAINER_AUTHORIZED_KEYS=~/.ssh/authorized_keys in .env")
	}
}

func TestWriteEnvFile_ExistingKey_Updates(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	initialContent := "PORT=8080\nREX_CONTAINER_AUTHORIZED_KEYS=/old/key/path\nDEBUG=true\n"
	if err := os.WriteFile(envPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to create initial .env: %v", err)
	}

	err := WriteEnvFile(dir)
	if err != nil {
		t.Fatalf("expected WriteEnvFile to succeed, got: %v", err)
	}

	data, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("failed to read .env: %v", err)
	}

	content := string(data)
	if strings.Contains(content, "/old/key/path") {
		t.Errorf("expected old key path to be overwritten")
	}
	if !strings.Contains(content, "REX_CONTAINER_AUTHORIZED_KEYS=~/.ssh/authorized_keys") {
		t.Errorf("expected updated REX_CONTAINER_AUTHORIZED_KEYS in .env")
	}
	if !strings.Contains(content, "PORT=8080") || !strings.Contains(content, "DEBUG=true") {
		t.Errorf("expected other vars to be preserved")
	}
}

func TestWriteEnvFile_WithoutTrailingNewline(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	initialContent := "PORT=8080"
	if err := os.WriteFile(envPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to create initial .env: %v", err)
	}

	err := WriteEnvFile(dir)
	if err != nil {
		t.Fatalf("expected WriteEnvFile to succeed, got: %v", err)
	}

	data, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("failed to read .env: %v", err)
	}

	content := string(data)
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d in: %s", len(lines), content)
	}
	if lines[0] != "PORT=8080" {
		t.Errorf("expected first line 'PORT=8080', got '%s'", lines[0])
	}
	if lines[1] != "REX_CONTAINER_AUTHORIZED_KEYS=~/.ssh/authorized_keys" {
		t.Errorf("expected second line 'REX_CONTAINER_AUTHORIZED_KEYS=~/.ssh/authorized_keys', got '%s'", lines[1])
	}
}
