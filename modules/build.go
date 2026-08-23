package modules

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"rexctl/logging"
)

// CmdBuild builds container images without starting them (docker compose build).
func CmdBuild(args []string) {
	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}
	target, stackDir := ResolveWorkspaceOrDie(arg)

	ValidateStack(stackDir)
	m := parseManifest(stackDir)

	logging.LogInfo("Building images for workspace '%s'...", target)

	for _, c := range m.Spec.Containers {
		if c.Type == "compose" {
			contDir := filepath.Join(stackDir, c.Name)
			if _, err := os.Stat(contDir); os.IsNotExist(err) {
				Die("Container '%s' is missing files. Run 'rexctl sync %s' first.", c.Name, target)
			}

			if err := PrepareComposePrebuiltImages(contDir, target, false); err != nil {
				Die("Failed to prepare pre-built images for '%s': %v", c.Name, err)
			}

			// Refresh standard override with latest commit and dirty state before building
			rev, _ := GetGitCommitHash(contDir)
			dirty, _ := IsGitDirty(contDir)
			if err := WriteStandardOverride(contDir, target, c.Name, rev, dirty); err != nil {
				logging.LogWarn("Failed to update docker-compose.override.yml in '%s': %v", c.Name, err)
			}

			composeArgs := BuildComposeArgs(target, "build")
			cmd := exec.Command("docker", composeArgs...)

			cmd.Dir = contDir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				Die("Failed to build compose container '%s'. Docker compose build exited with an error.", c.Name)
			}

		} else if c.Type == "image" {
			imageTag := fmt.Sprintf("rexctl/%s/%s:latest", target, c.Name)
			if err := exec.Command("docker", "image", "inspect", imageTag).Run(); err != nil {
				logging.LogInfo("Pulling and tagging remote image for '%s'...", c.Name)
				if _, err := runCmd(stackDir, "docker", "pull", c.Remote); err != nil {
					Die("Failed to pull image '%s': %v", c.Remote, err)
				}
				if _, err := runCmd(stackDir, "docker", "tag", c.Remote, imageTag); err != nil {
					Die("Failed to tag image '%s': %v", c.Remote, err)
				}
			} else {
				logging.LogInfo("Image '%s' (%s) already exists locally.", c.Name, imageTag)
			}
		}
	}

	logging.LogInfo("Images for workspace '%s' built successfully.", target)
}
