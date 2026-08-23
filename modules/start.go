package modules

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"rexctl/logging"
	"strings"
)

// CmdUp brings up containers (docker compose up -d) letting docker compose automatically discover compose files.
func CmdUp(args []string) {
	if len(args) < 1 {
		Die("Usage: rexctl up <stack>")
	}
	target, stackDir := ResolveWorkspaceOrDie(args[0])

	active, found := getActiveStack()
	if found {
		if active == target {
			logging.LogInfo("Stack '%s' is already active. Ensuring all containers are up...", target)
		} else {
			Die("Workspace '%s' is currently running! Use 'rexctl switch %s' to change workspaces.", active, target)
		}
	}

	ValidateStack(stackDir)
	m := parseManifest(stackDir)

	for _, c := range m.Spec.Containers {
		if c.Type == "compose" {
			contDir := filepath.Join(stackDir, c.Name)
			if _, err := os.Stat(contDir); os.IsNotExist(err) {
				Die("Container '%s' is missing files. Run 'rexctl sync %s' first.", c.Name, target)
			}

			if err := PrepareComposePrebuiltImages(contDir, target, false); err != nil {
				Die("Failed to prepare pre-built images for '%s': %v", c.Name, err)
			}

			// Refresh standard override with latest commit and dirty state before bringing up
			rev, _ := GetGitCommitHash(contDir)
			dirty, _ := IsGitDirty(contDir)
			if err := WriteStandardOverride(contDir, target, c.Name, rev, dirty); err != nil {
				logging.LogWarn("Failed to update docker-compose.override.yml in '%s': %v", c.Name, err)
			}

			composeArgs := BuildComposeArgs(target, "up", "-d", "--build")
			cmd := exec.Command("docker", composeArgs...)

			cmd.Dir = contDir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				Die("Failed to bring up compose container '%s'. Docker compose exited with an error.", c.Name)
			}

		} else if c.Type == "image" {
			containerName := fmt.Sprintf("%s-%s", c.Name, target)
			imageTag := fmt.Sprintf("rexctl/%s/%s:latest", target, c.Name)

			if err := exec.Command("docker", "image", "inspect", imageTag).Run(); err != nil {
				Die("Image '%s' not found locally. Run 'rexctl sync %s' first.", imageTag, target)
			}

			// Recreate image container to ensure latest image and network attachment
			exec.Command("docker", "rm", "-f", containerName).Run()

			runArgs := []string{
				"run", "-d",
				"--name", containerName,
				"--label", fmt.Sprintf("com.docker.compose.project=%s", target),
				"--label", fmt.Sprintf("rexctl.workspace=%s", target),
				"--label", fmt.Sprintf("rexctl.service=%s", c.Name),
			}

			// Attach to compose network if it exists
			netCheck := exec.Command("docker", "network", "inspect", fmt.Sprintf("%s_default", target))
			if netCheck.Run() == nil {
				runArgs = append(runArgs, "--network", fmt.Sprintf("%s_default", target))
			}

			runArgs = append(runArgs, imageTag)

			runCmd := exec.Command("docker", runArgs...)
			runCmd.Stdout = os.Stdout
			runCmd.Stderr = os.Stderr

			if err := runCmd.Run(); err != nil {
				Die("Failed to create image container '%s'. Docker run exited with an error.", c.Name)
			}
		}
	}

	logging.LogInfo("Stack '%s' is up.", target)
}

// CmdStart starts existing, stopped containers (docker compose start).
func CmdStart(args []string) {
	if len(args) < 1 {
		Die("Usage: rexctl start <stack>")
	}
	target, stackDir := ResolveWorkspaceOrDie(args[0])

	active, found := getActiveStack()
	if found {
		if active == target {
			logging.LogInfo("Stack '%s' is already active.", target)
		} else {
			Die("Workspace '%s' is currently running! Use 'rexctl switch %s' to change workspaces.", active, target)
		}
	}

	ValidateStack(stackDir)
	m := parseManifest(stackDir)

	for _, c := range m.Spec.Containers {
		if c.Type == "compose" {
			contDir := filepath.Join(stackDir, c.Name)
			if _, err := os.Stat(contDir); os.IsNotExist(err) {
				Die("Container '%s' is missing files. Run 'rexctl sync %s' first.", c.Name, target)
			}

			composeArgs := BuildComposeArgs(target, "start")
			cmd := exec.Command("docker", composeArgs...)
			cmd.Dir = contDir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				psCmd := exec.Command("docker", "compose", "--project-name", target, "ps", "-a", "-q")
				psCmd.Dir = contDir
				out, _ := psCmd.Output()
				if len(strings.TrimSpace(string(out))) == 0 {
					Die("No existing containers found for compose service '%s'. Run 'rexctl up %s' first to build and create the containers.", c.Name, target)
				}
				Die("Failed to start compose container '%s'. Docker compose exited with an error.", c.Name)
			}

		} else if c.Type == "image" {
			containerName := fmt.Sprintf("%s-%s", c.Name, target)
			startCmd := exec.Command("docker", "start", containerName)
			startCmd.Stdout = os.Stdout
			startCmd.Stderr = os.Stderr
			if err := startCmd.Run(); err != nil {
				Die("Failed to start image container '%s'. Run 'rexctl up %s' first if container does not exist.", c.Name, target)
			}
		}
	}

	logging.LogInfo("Stack '%s' started.", target)
}
