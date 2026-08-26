package modules

import (
	"fmt"
	"rexctl/config"
	"strings"
)

// GetGitCommitHash returns the short HEAD commit hash of a git repository.
func GetGitCommitHash(repoDir string) (string, error) {
	out, err := runCmd(repoDir, "git", "rev-parse", "--short", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// IsGitDirty returns true if the git working tree has staged, unstaged, or untracked changes
// (ignoring rexctl-generated files like docker-compose.override.yml, .env, and init script).
func IsGitDirty(repoDir string) (bool, error) {
	out, err := runCmd(repoDir, "git", "status", "--porcelain")
	if err != nil {
		return false, err
	}
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Ignore rexctl-generated files (override files, .env, init scripts)
		if isIgnoredGitFile(trimmed) {
			continue
		}
		return true, nil
	}
	return false, nil
}

// isIgnoredGitFile checks whether a git status line corresponds to an ignored rexctl file (.env, docker compose overrides, or init script).
func isIgnoredGitFile(trimmed string) bool {
	if strings.HasSuffix(trimmed, "docker-compose.override.yml") ||
		strings.HasSuffix(trimmed, "compose.override.yaml") ||
		strings.HasSuffix(trimmed, "docker-compose.override.yaml") ||
		strings.HasSuffix(trimmed, " .env") ||
		strings.HasSuffix(trimmed, "/.env") ||
		strings.HasSuffix(trimmed, "\t.env") ||
		trimmed == ".env" {
		return true
	}
	if config.DefaultInitScriptName != "" &&
		(strings.HasSuffix(trimmed, " "+config.DefaultInitScriptName) ||
			strings.HasSuffix(trimmed, "/"+config.DefaultInitScriptName) ||
			strings.HasSuffix(trimmed, "\t"+config.DefaultInitScriptName) ||
			trimmed == config.DefaultInitScriptName) {
		return true
	}
	return false
}



// BuildImageTag creates a standardized image tag including workspace, container name, commit hash, and dirty suffix.
// Format: rexctl/<workspace>/<containerName>:<commit_hash>[-dirty]
func BuildImageTag(workspace, containerName, commitHash string, isDirty bool) string {
	if commitHash != "" {
		tag := commitHash
		if isDirty {
			tag += "-dirty"
		}
		return fmt.Sprintf("rexctl/%s/%s:%s", workspace, containerName, tag)
	}
	return fmt.Sprintf("rexctl/%s/%s:latest", workspace, containerName)
}
