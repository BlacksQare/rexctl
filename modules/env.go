package modules

import (
	"fmt"
	"os"
	"path/filepath"
	"rexctl/config"
	"strings"
)

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	} else if path == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			return home
		}
	}
	return path
}

// AuthorizedKeysExists checks whether the host contains the configured authorized_keys file.
func AuthorizedKeysExists() bool {
	if config.DefaultAuthorizedKeysPath == "" {
		return false
	}
	expanded := expandHome(config.DefaultAuthorizedKeysPath)
	if _, err := os.Stat(expanded); err == nil {
		return true
	}
	return false
}

// WriteEnvFile creates or updates a .env file in repoDir containing the REX_CONTAINER_AUTHORIZED_KEYS variable,
// but only if the authorized_keys file exists on the host.
func WriteEnvFile(repoDir string) error {
	key := config.DefaultEnvVarAuthorizedKeys
	envPath := filepath.Join(repoDir, ".env")

	if !AuthorizedKeysExists() {
		// If authorized_keys does not exist on host, do not add it.
		// If .env exists and contains REX_CONTAINER_AUTHORIZED_KEYS, remove that entry.
		data, err := os.ReadFile(envPath)
		if err != nil {
			return nil // No .env file, nothing to do
		}
		lines := strings.Split(string(data), "\n")
		var newLines []string
		changed := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, key+"=") || trimmed == key {
				changed = true
				continue
			}
			newLines = append(newLines, line)
		}
		if changed {
			content := strings.Join(newLines, "\n")
			if strings.TrimSpace(content) == "" {
				return os.Remove(envPath)
			}
			return os.WriteFile(envPath, []byte(content), 0644)
		}
		return nil
	}

	val := config.DefaultAuthorizedKeysPath
	entry := fmt.Sprintf("%s=%s", key, val)

	data, err := os.ReadFile(envPath)
	if err != nil {
		if os.IsNotExist(err) {
			return os.WriteFile(envPath, []byte(entry+"\n"), 0644)
		}
		return err
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	found := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, key+"=") || trimmed == key {
			lines[i] = entry
			found = true
			break
		}
	}

	if !found {
		if len(content) > 0 && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += entry + "\n"
		return os.WriteFile(envPath, []byte(content), 0644)
	}

	return os.WriteFile(envPath, []byte(strings.Join(lines, "\n")), 0644)
}
