package modules

import (
	"os"
	"path/filepath"
	"rexctl/config"
	"strings"
	"testing"
)

func setupDummyAuthorizedKeys(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "authorized_keys")
	if err := os.WriteFile(keyPath, []byte("ssh-rsa AAAAB3NzaC1yc2E dummy"), 0600); err != nil {
		t.Fatalf("failed to create dummy authorized_keys: %v", err)
	}
	orig := config.DefaultAuthorizedKeysPath
	config.DefaultAuthorizedKeysPath = keyPath
	t.Cleanup(func() {
		config.DefaultAuthorizedKeysPath = orig
	})
	return keyPath
}

func TestWriteEnvFile_WhenKeyExists_CreatesNewFile(t *testing.T) {
	keyPath := setupDummyAuthorizedKeys(t)
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
	expected := "REX_CONTAINER_AUTHORIZED_KEYS=" + keyPath + "\n"
	if content != expected {
		t.Errorf("expected '%s', got '%s'", expected, content)
	}
}

func TestWriteEnvFile_WhenKeyDoesNotExist_DoesNotCreateFile(t *testing.T) {
	orig := config.DefaultAuthorizedKeysPath
	config.DefaultAuthorizedKeysPath = "/nonexistent/path/to/authorized_keys"
	t.Cleanup(func() {
		config.DefaultAuthorizedKeysPath = orig
	})

	dir := t.TempDir()
	err := WriteEnvFile(dir)
	if err != nil {
		t.Fatalf("expected WriteEnvFile to succeed, got: %v", err)
	}

	envPath := filepath.Join(dir, ".env")
	if _, err := os.Stat(envPath); !os.IsNotExist(err) {
		t.Errorf("expected .env NOT to exist when authorized_keys is missing, but it exists")
	}
}

func TestWriteEnvFile_WhenKeyDoesNotExist_RemovesExistingKeyFromEnv(t *testing.T) {
	orig := config.DefaultAuthorizedKeysPath
	config.DefaultAuthorizedKeysPath = "/nonexistent/path/to/authorized_keys"
	t.Cleanup(func() {
		config.DefaultAuthorizedKeysPath = orig
	})

	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	initialContent := "PORT=8080\nREX_CONTAINER_AUTHORIZED_KEYS=/some/old/path\nDEBUG=true\n"
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
	if strings.Contains(content, "REX_CONTAINER_AUTHORIZED_KEYS") {
		t.Errorf("expected REX_CONTAINER_AUTHORIZED_KEYS to be removed from .env when key does not exist on host")
	}
	if !strings.Contains(content, "PORT=8080") || !strings.Contains(content, "DEBUG=true") {
		t.Errorf("expected other vars to be preserved, got: %s", content)
	}
}

func TestWriteEnvFile_ExistingFile_Appends(t *testing.T) {
	keyPath := setupDummyAuthorizedKeys(t)
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
	if !strings.Contains(content, "REX_CONTAINER_AUTHORIZED_KEYS="+keyPath) {
		t.Errorf("expected REX_CONTAINER_AUTHORIZED_KEYS=%s in .env, got: %s", keyPath, content)
	}
}

func TestWriteEnvFile_ExistingKey_Updates(t *testing.T) {
	keyPath := setupDummyAuthorizedKeys(t)
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
	if !strings.Contains(content, "REX_CONTAINER_AUTHORIZED_KEYS="+keyPath) {
		t.Errorf("expected updated REX_CONTAINER_AUTHORIZED_KEYS in .env")
	}
	if !strings.Contains(content, "PORT=8080") || !strings.Contains(content, "DEBUG=true") {
		t.Errorf("expected other vars to be preserved")
	}
}

func TestWriteEnvFile_WithoutTrailingNewline(t *testing.T) {
	keyPath := setupDummyAuthorizedKeys(t)
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
	if lines[1] != "REX_CONTAINER_AUTHORIZED_KEYS="+keyPath {
		t.Errorf("expected second line 'REX_CONTAINER_AUTHORIZED_KEYS=%s', got '%s'", keyPath, lines[1])
	}
}

