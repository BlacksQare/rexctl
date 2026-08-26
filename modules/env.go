package modules

import (
	"fmt"
	"os"
	"path/filepath"
	"rexctl/config"
	"strings"
)

// WriteEnvFile creates or updates a .env file in repoDir containing the REX_CONTAINER_AUTHORIZED_KEYS variable.
func WriteEnvFile(repoDir string) error {
	envPath := filepath.Join(repoDir, ".env")
	key := config.DefaultEnvVarAuthorizedKeys
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
