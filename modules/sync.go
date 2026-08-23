package modules

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"rexctl/logging"
	"rexctl/structs"
	"strings"
	"sync"
)

func askForConfirmation(prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

func CmdSync(args []string) {
	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}
	stackName, stackDir := ResolveWorkspaceOrDie(arg)

	if _, err := os.Stat(filepath.Join(stackDir, "rex.yaml")); err != nil {
		Die("Stack manifest not found: %s\nRun 'rexctl create %s' first.", filepath.Join(stackDir, "rex.yaml"), stackName)
	}

	logging.LogInfo("------------------------------------------------------------")
	logging.LogInfo("WARNING: Syncing will rebuild container states for '%s'.", stackName)
	logging.LogInfo("Existing containers will be removed and recreated from the remote repository.")
	logging.LogInfo("------------------------------------------------------------")

	if !askForConfirmation("Are you sure you want to proceed with sync?") {
		logging.LogInfo("Sync aborted.")
		return
	}

	m := parseManifest(stackDir)

	for _, c := range m.Spec.Containers {
		if c.Type == "compose" {
			contDir := filepath.Join(stackDir, c.Name)
			if _, err := os.Stat(contDir); err == nil {
				dirty, err := IsGitDirty(contDir)
				if err != nil {
					Die("Failed to check status for '%s': %v", c.Name, err)
				}
				if dirty {
					logging.LogInfo("WARNING: Container '%s' has staged, unstaged, or untracked changes.", c.Name)
					if askForConfirmation("Force recreate (destroy local changes and re-clone)?") {
						if err := os.RemoveAll(contDir); err != nil {
							Die("Failed to remove '%s': %v", c.Name, err)
						}
						logging.LogInfo("Local changes destroyed. Repository will be cloned fresh.")
					} else {
						Die("Aborting sync.")
					}
				}
			}
		}
	}

	var wg sync.WaitGroup
	var syncErrors []error
	var syncMu sync.Mutex

	for _, c := range m.Spec.Containers {
		wg.Add(1)
		go func(c structs.Container) {
			defer wg.Done()
			if c.Type == "compose" {
				contDir := filepath.Join(stackDir, c.Name)

				if _, err := os.Stat(contDir); err != nil {
					logging.LogInfo("Cloning '%s'...", c.Name)
					if _, err := runCmd(stackDir, "git", "clone", "--recurse-submodules", c.Remote, c.Name); err != nil {
						syncMu.Lock()
						syncErrors = append(syncErrors, fmt.Errorf("failed to clone '%s': %w", c.Name, err))
						syncMu.Unlock()
						return
					}
					if _, err := runCmd(contDir, "git", "checkout", c.Revision); err != nil {
						syncMu.Lock()
						syncErrors = append(syncErrors, fmt.Errorf("failed to checkout '%s' in '%s': %w", c.Revision, c.Name, err))
						syncMu.Unlock()
						return
					}
					runCmd(contDir, "git", "submodule", "update", "--init", "--recursive")
				} else {
					logging.LogInfo("Fetching and checking out '%s'...", c.Name)
					runCmd(contDir, "git", "fetch", "--all", "--tags")
					if _, err := runCmd(contDir, "git", "checkout", c.Revision); err != nil {
						syncMu.Lock()
						syncErrors = append(syncErrors, fmt.Errorf("failed to checkout '%s' in '%s': %w", c.Revision, c.Name, err))
						syncMu.Unlock()
						return
					}
					runCmd(contDir, "git", "pull", "--ff-only")
					runCmd(contDir, "git", "submodule", "update", "--init", "--recursive")
				}

				if err := PrepareComposePrebuiltImages(contDir, stackName, true); err != nil {
					logging.LogWarn("Failed to prepare pre-built images for '%s': %v", c.Name, err)
				}

				rev, _ := GetGitCommitHash(contDir)
				dirty, _ := IsGitDirty(contDir)
				if err := WriteStandardOverride(contDir, stackName, c.Name, rev, dirty); err != nil {
					logging.LogWarn("Failed to write docker-compose.override.yml in '%s': %v", c.Name, err)
				}

			} else if c.Type == "image" {
				logging.LogInfo("Pulling registry image for '%s'...", c.Name)
				if _, err := runCmd(stackDir, "docker", "pull", c.Remote); err != nil {
					syncMu.Lock()
					syncErrors = append(syncErrors, fmt.Errorf("failed to pull image '%s': %w", c.Remote, err))
					syncMu.Unlock()
					return
				}
				if _, err := runCmd(stackDir, "docker", "tag", c.Remote, fmt.Sprintf("rexctl/%s/%s:latest", stackName, c.Name)); err != nil {
					syncMu.Lock()
					syncErrors = append(syncErrors, fmt.Errorf("failed to tag image '%s': %w", c.Remote, err))
					syncMu.Unlock()
					return
				}
			}
		}(c)
	}

	wg.Wait()

	if len(syncErrors) > 0 {
		for _, err := range syncErrors {
			logging.LogErr("%v", err)
		}
		Die("Sync failed with %d error(s).", len(syncErrors))
	}

	ValidateStack(stackDir)

	logging.LogInfo("Recreating container states for '%s'...", stackName)
	for _, c := range m.Spec.Containers {
		if c.Type == "compose" {
			contDir := filepath.Join(stackDir, c.Name)
			if _, err := os.Stat(contDir); err == nil {
				cmd := exec.Command("docker", "compose", "--project-name", stackName, "down")
				cmd.Dir = contDir
				cmd.Run()
			}
		} else if c.Type == "image" {
			containerName := fmt.Sprintf("%s-%s", c.Name, stackName)
			exec.Command("docker", "rm", "-f", containerName).Run()
		}
	}

	if active, found := getActiveStack(); found && active == stackName {
		logging.LogInfo("Stack is currently active, applying new state...")
		CmdUp([]string{stackName})
	}

	logging.LogInfo("Sync complete.")
}
