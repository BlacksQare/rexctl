package modules

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"rexctl/config"
	"rexctl/logging"
)

func CmdStart(args []string) {
	if len(args) < 1 {
		Die("Usage: rexctl start <stack>")
	}
	target := args[0]

	active, found := getActiveStack()
	if found {
		if active == target {
			logging.LogInfo("Stack '%s' is already active. Ensuring all containers are up...", target)
		} else {
			Die("Workspace '%s' is currently running! Use 'rexctl switch %s' to change workspaces.", active, target)
		}
	}

	stackDir := filepath.Join(config.WorkspacesDir, target)

	ValidateStack(stackDir)
	m := parseManifest(stackDir)

	for _, c := range m.Spec.Containers {
		if c.Type == "compose" {
			relComposeFile := filepath.Join(c.Name, c.ComposeFile)
			absComposeFile := filepath.Join(stackDir, relComposeFile)

			if _, err := os.Stat(absComposeFile); os.IsNotExist(err) {
				Die("Container '%s' is missing files. Run 'rexctl sync %s' first.", c.Name, target)
			}

			relOverrideFile := createComposeOverride(stackDir, relComposeFile, target)

			composeArgs := []string{"compose", "--project-name", target, "-f", relComposeFile}
			if relOverrideFile != "" {
				composeArgs = append(composeArgs, "-f", relOverrideFile)
			}
			composeArgs = append(composeArgs, "up", "-d")

			cmd := exec.Command("docker", composeArgs...)
			cmd.Dir = stackDir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				Die("Failed to start compose container '%s'. Docker compose exited with an error.", c.Name)
			}

		} else if c.Type == "image" {
			containerName := fmt.Sprintf("%s-%s", c.Name, target)
			imageTag := fmt.Sprintf("rex/%s/%s", target, c.Name)

			if err := exec.Command("docker", "image", "inspect", imageTag).Run(); err != nil {
				Die("Image '%s' not found locally. Run 'rexctl sync %s' first.", imageTag, target)
			}

			startCmd := exec.Command("docker", "start", containerName)
			if err := startCmd.Run(); err != nil {
				runCmd := exec.Command("docker", "run", "-d",
					"--name", containerName,
					"--label", fmt.Sprintf("com.docker.compose.project=%s", target),
					imageTag)

				runCmd.Stdout = os.Stdout
				runCmd.Stderr = os.Stderr

				if err := runCmd.Run(); err != nil {
					Die("Failed to create image container '%s'. Docker run exited with an error.", c.Name)
				}
			}
		}
	}

	logging.LogInfo("Stack '%s' started.", target)
}

func CmdStop() {
	active, found := getActiveStack()
	if !found {
		Die("No active stack running.")
	}

	logging.LogInfo("Stopping stack '%s'...", active)
	stackDir := filepath.Join(config.WorkspacesDir, active)

	m := parseManifest(stackDir)
	for _, repo := range m.Spec.Containers {
		ComposeFile := filepath.Join(repo.Name, "docker-compose.yaml")
		if _, err := os.Stat(filepath.Join(stackDir, ComposeFile)); err == nil {
			runCmdStream(stackDir, "docker", "compose", "--project-name", active, "-f", ComposeFile, "stop")
		}
	}
	logging.LogInfo("Stack '%s' stopped.", active)
}
