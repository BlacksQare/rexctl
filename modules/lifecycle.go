package modules

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"rexctl/config"
	"rexctl/logging"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
)

type WorkspaceManifest struct {
	Kind string        `yaml:"kind"`
	Spec WorkspaceSpec `yaml:"spec"`
}

type WorkspaceSpec struct {
	Repositories []Repository `yaml:"repositories"`
}

type Repository struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Remote   string `yaml:"remote"`
	Revision string `yaml:"revision,omitempty"`
}

func CmdCreate(args []string) {
	if len(args) < 1 {
		Die("Usage: rexctl create <stack>")
	}

	stackName := args[0]
	stackDir := filepath.Join(config.WorkspacesDir, stackName)
	var manifestData []byte

	if config.DefaultManifestPath != "" {
		data, err := os.ReadFile(config.DefaultManifestPath)
		if err == nil {
			manifestData = data
		} else {
			Die("Failed to read default manifest from Nix store: %v", err)
		}
	} else {
		manifestData = []byte(config.DefaultManifestFallback)
	}

	if err := os.MkdirAll(stackDir, 0755); err != nil {
		Die("Failed to create workspace directory: %v", err)
	}

	targetFile := filepath.Join(stackDir, "rex.yaml")
	if err := os.WriteFile(targetFile, manifestData, 0644); err != nil {
		Die("Failed to create rex.yaml: %v", err)
	}

	logging.LogInfo("Workspace '%s' initialized.", stackName)
}

func CmdPwd(args []string) {
	if len(args) < 1 {
		Die("Usage: rexctl pwd <stack>")
	}
	fmt.Println(filepath.Join(config.WorkspacesDir, args[0]))
}

func CmdGet() {
	active, found := getActiveStack()
	if found && active != "" {
		fmt.Println(active)
	} else {
		fmt.Println("No workspace is currently running.")
	}
}

func CmdStatus() {
	active, found := getActiveStack()
	if !found {
		fmt.Println("No active stack")
		return
	}

	cli := getDockerClient()
	f := filters.NewArgs()
	f.Add("label", "com.docker.compose.project="+active)
	containers, _ := cli.ContainerList(context.Background(), container.ListOptions{All: true, Filters: f})

	running := 0
	for _, c := range containers {
		if c.State == "running" {
			running++
		}
	}

	fmt.Printf("%d out of %d containers running from stack %s\n", running, len(containers), active)
}

func CmdDown(args []string) {
	var target string

	if len(args) > 0 {
		target = args[0]
	} else {
		active, found := getActiveStack()
		if !found {
			Die("No active stack running to bring down.")
		}
		target = active
	}

	logging.LogInfo("Stopping stack '%s'...", target)
	stackDir := filepath.Join(config.WorkspacesDir, target)

	if _, err := os.Stat(filepath.Join(stackDir, "rex.yaml")); err == nil {
		m := parseManifest(stackDir)
		for _, c := range m.Spec.Containers {
			if c.Type == "compose" {
				relComposeFile := filepath.Join(c.Name, c.ComposeFile)
				absComposeFile := filepath.Join(stackDir, relComposeFile)

				if _, err := os.Stat(absComposeFile); err == nil {
					relOverrideFile := createComposeOverride(stackDir, relComposeFile, target)

					composeArgs := []string{"compose", "--project-name", target, "-f", relComposeFile}
					if relOverrideFile != "" {
						composeArgs = append(composeArgs, "-f", relOverrideFile)
					}
					composeArgs = append(composeArgs, "stop")

					cmd := exec.Command("docker", composeArgs...)
					cmd.Dir = stackDir
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					cmd.Run()

					if relOverrideFile != "" {
						os.Remove(filepath.Join(stackDir, relOverrideFile))
					}
				}
			}
		}
	}

	out, err := exec.Command("docker", "ps", "-q", "--filter", fmt.Sprintf("label=com.docker.compose.project=%s", target)).Output()
	if err == nil {
		containerIDs := strings.Fields(string(out))
		if len(containerIDs) > 0 {
			stopArgs := append([]string{"stop"}, containerIDs...)
			stopCmd := exec.Command("docker", stopArgs...)
			stopCmd.Stdout = os.Stdout
			stopCmd.Stderr = os.Stderr
			stopCmd.Run()
		}
	}

	logging.LogInfo("Stack '%s' stopped.", target)
}

func CmdSwitch(args []string) {
	if len(args) < 1 {
		Die("Usage: rexctl switch <stack>")
	}
	target := args[0]
	ValidateStack(filepath.Join(config.WorkspacesDir, target))

	if active, found := getActiveStack(); found {
		if active == target {
			logging.LogInfo("Stack '%s' is already running.", target)
			return
		}
		CmdDown([]string{active})
	}

	CmdStart([]string{target})
}

func CmdDestroy(args []string) {
	if len(args) < 1 {
		Die("Usage: rexctl destroy <stack>")
	}
	target := args[0]
	stackDir := filepath.Join(config.WorkspacesDir, target)

	if _, err := os.Stat(stackDir); os.IsNotExist(err) {
		Die("Stack '%s' does not exist.", target)
	}

	logging.LogInfo("------------------------------------------------------------")
	logging.LogInfo("WARNING: You are about to permanently destroy workspace '%s'.", target)
	logging.LogInfo("This will delete all cloned repositories, manifests, and local data.")
	logging.LogInfo("------------------------------------------------------------")

	if !askForConfirmation("Are you absolutely sure you want to proceed?") {
		logging.LogInfo("Destroy aborted.")
		return
	}

	if active, found := getActiveStack(); found && active == target {
		CmdDown([]string{target})
	}

	logging.LogInfo("Destroying stack '%s'...", target)

	if err := forceRemoveDir(stackDir); err != nil {
		Die("Failed to remove stack directory: %v", err)
	}

	logging.LogInfo("Stack '%s' destroyed.", target)
}

func CmdEdit(args []string) {
	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}
	stackName := getTargetStack(arg)
	stackDir := filepath.Join(config.WorkspacesDir, stackName)
	targetFile := filepath.Join(stackDir, "rex.yaml")

	if _, err := os.Stat(targetFile); err != nil {
		Die("Stack manifest not found: %s\nRun 'rexctl create %s' first.", targetFile, stackName)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	cmd := exec.Command(editor, targetFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		Die("Editor '%s' failed or exited with an error: %v", editor, err)
	}

	logging.LogInfo("Validating manifest...")

	ValidateStack(stackDir)

	logging.LogInfo("Manifest is properly validated.")

	if active, found := getActiveStack(); found && active == stackName {
		logging.LogInfo("------------------------------------------------------------")
		logging.LogInfo("NOTICE: Workspace '%s' is currently running.", stackName)
		logging.LogInfo("To apply these changes, you must sync and restart the stack:")
		logging.LogInfo("  rexctl sync %s", stackName)
		logging.LogInfo("  rexctl switch %s   (or rexctl down && rexctl start)", stackName)
		logging.LogInfo("------------------------------------------------------------")
	}
}

func Die(format string, a ...any) {
	logging.LogErr(format, a...)
	os.Exit(1)
}
