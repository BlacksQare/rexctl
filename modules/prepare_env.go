package modules

import (
	"fmt"
	"os"
	"path/filepath"
	"rexctl/config"
	"rexctl/logging"
)

// RunPrepareEnv executes the workspace init script (e.g. rexctl_init.sh) in every repository in the workspace.
// It iterates through all containers/repos defined in the manifest and executes the script if present.
func RunPrepareEnv(stackDir string) error {
	m, err := ParseManifest(stackDir)
	if err != nil {
		return fmt.Errorf("failed to parse workspace manifest: %w", err)
	}

	for _, c := range m.Spec.Containers {
		if c.Type != "compose" {
			continue
		}
		repoDir := filepath.Join(stackDir, c.Name)
		if err := WriteEnvFile(repoDir); err != nil {
			logging.LogWarn("Failed to write .env in '%s': %v", c.Name, err)
		}
		initScript := filepath.Join(repoDir, config.DefaultInitScriptName)

		if _, err := os.Stat(initScript); err == nil {
			logging.LogInfo("Executing '%s' in repository '%s'...", config.DefaultInitScriptName, c.Name)
			if err := runCmdStream(repoDir, "bash", config.DefaultInitScriptName); err != nil {
				return fmt.Errorf("init script '%s' failed in repo '%s': %w", config.DefaultInitScriptName, c.Name, err)
			}
			logging.LogInfo("Completed '%s' in repository '%s'.", config.DefaultInitScriptName, c.Name)
		} else {
			logging.LogInfo("No '%s' found in '%s' (skipping).", config.DefaultInitScriptName, c.Name)
		}
	}

	return nil
}

// CmdPrepareEnv is the CLI entry point for 'rexctl prepare-env [workspace]'.
func CmdPrepareEnv(args []string) {
	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}
	stackName, stackDir := ResolveWorkspaceOrDie(arg)

	if _, err := os.Stat(filepath.Join(stackDir, "rex.yaml")); err != nil {
		Die("Stack manifest not found: %s\nRun 'rexctl create %s' first.", filepath.Join(stackDir, "rex.yaml"), stackName)
	}

	logging.LogInfo("Preparing environment for workspace '%s'...", stackName)
	if err := RunPrepareEnv(stackDir); err != nil {
		Die("prepare-env failed: %v", err)
	}
	logging.LogInfo("Environment preparation complete for workspace '%s'.", stackName)
}
