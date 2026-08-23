package modules

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"rexctl/config"
	"strings"
)

// ParseActiveStackFromOutput parses the active project name from docker ps output.
func ParseActiveStackFromOutput(out string) (string, bool) {
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			parts := strings.Split(line, "|")
			if len(parts) > 0 && parts[0] != "" {
				return parts[0], true
			} else if len(parts) > 1 && parts[1] != "" {
				return parts[1], true
			}
			return line, true
		}
	}
	return "", false
}

func getActiveStack() (string, bool) {
	cmd := exec.Command("docker", "ps",
		"--filter", "status=running",
		"--format", "{{.Label \"rexctl.workspace\"}}|{{.Label \"com.docker.compose.project\"}}")

	out, err := cmd.Output()
	if err != nil {
		return "", false
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		wsLabel := ""
		composeProject := ""
		if len(parts) > 0 {
			wsLabel = strings.TrimSpace(parts[0])
		}
		if len(parts) > 1 {
			composeProject = strings.TrimSpace(parts[1])
		}

		if wsLabel != "" {
			return wsLabel, true
		}
		if composeProject != "" {
			// Verify this compose project is a known rexctl workspace
			wsDir := filepath.Join(config.WorkspacesDir, composeProject)
			if _, statErr := os.Stat(filepath.Join(wsDir, "rex.yaml")); statErr == nil {
				return composeProject, true
			}
			if cwd, err := os.Getwd(); err == nil && filepath.Base(cwd) == composeProject {
				if _, statErr := os.Stat(filepath.Join(cwd, "rex.yaml")); statErr == nil {
					return composeProject, true
				}
			}
		}
	}

	return "", false
}

// GetActiveStack returns the active stack name if at least one container is running.
func GetActiveStack() (string, bool) {
	return getActiveStack()
}

// ResolveRunningContainer finds the running container name or ID for a given service/container identifier.
func ResolveRunningContainer(stackName, containerName string) string {
	cmd := exec.Command("docker", "ps",
		"--filter", "status=running",
		"--format", "{{.ID}}|{{.Names}}|{{.Label \"rexctl.service\"}}|{{.Label \"com.docker.compose.service\"}}|{{.Label \"com.docker.compose.project\"}}|{{.Label \"rexctl.workspace\"}}")

	out, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			fields := strings.Split(line, "|")
			if len(fields) < 6 {
				continue
			}
			id, names, rexSvc, composeSvc, composeProj, rexWs := fields[0], fields[1], fields[2], fields[3], fields[4], fields[5]

			projMatches := (stackName == "" || composeProj == stackName || rexWs == stackName)

			if projMatches {
				if rexSvc == containerName || composeSvc == containerName || names == containerName || id == containerName {
					return names
				}
				// Check if container name ends with -<containerName>-1 or _<containerName>_1 or starts with <containerName>-
				for _, n := range strings.Split(names, ",") {
					if n == containerName ||
						n == fmt.Sprintf("%s-%s-1", stackName, containerName) ||
						n == fmt.Sprintf("%s_%s_1", stackName, containerName) ||
						n == fmt.Sprintf("%s-%s", containerName, stackName) ||
						n == fmt.Sprintf("%s-%s", stackName, containerName) {
						return n
					}
				}
			}
		}
	}

	// Fallback to containerName directly
	return containerName
}
