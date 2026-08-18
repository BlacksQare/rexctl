package modules

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"rexctl/config"
	"rexctl/logging"
	"strings"
)

// CreateWorkspace creates a workspace folder and writes its rex.yaml manifest.
func CreateWorkspace(name string, manifest []byte) error {
	dir := filepath.Join(config.WorkspacesDir, name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "rex.yaml"), manifest, 0644)
}

func CmdCreate(args []string) {
	if len(args) < 1 {
		Die("Usage: rexctl create <stack>")
	}
	name := args[0]
	dir := filepath.Join(config.WorkspacesDir, name)

	if _, err := os.Stat(dir); err == nil {
		Die("Directory '%s' already exists for workspace '%s'.", dir, name)
	}

	var manifestContent []byte
	if config.DefaultManifestPath != "" {
		data, err := os.ReadFile(config.DefaultManifestPath)
		if err == nil {
			manifestContent = data
		}
	}
	if len(manifestContent) == 0 {
		manifestContent = []byte(config.DefaultManifestFallback)
	}

	if err := CreateWorkspace(name, manifestContent); err != nil {
		Die("Failed to create workspace directory or manifest: %v", err)
	}

	logging.LogInfo("Workspace '%s' created at %s", name, dir)
}

func CmdPwd(args []string) {
	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}
	_, stackDir := ResolveWorkspaceOrDie(arg)
	fmt.Println(stackDir)
}


func CmdGet() {
	active, found := getActiveStack()
	if !found || active == "" {
		return
	}
	fmt.Println(active)
}

func CmdStatus() {
	active, found := getActiveStack()
	if !found {
		fmt.Println("No active stack")
		return
	}

	outAll, _ := exec.Command("docker", "ps", "-a", "-q", "--filter", fmt.Sprintf("label=com.docker.compose.project=%s", active)).Output()
	allCount := len(strings.Fields(string(outAll)))

	outRunning, _ := exec.Command("docker", "ps", "-q", "--filter", fmt.Sprintf("label=com.docker.compose.project=%s", active)).Output()
	runningCount := len(strings.Fields(string(outRunning)))

	fmt.Printf("%d out of %d containers running from stack %s\n", runningCount, allCount, active)
}

// CmdStop gracefully stops running containers without removing them (similar to docker compose stop).
func CmdStop(args []string) {
	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}
	target, stackDir := ResolveWorkspaceOrDie(arg)

	logging.LogInfo("Stopping stack '%s'...", target)

	if _, err := os.Stat(filepath.Join(stackDir, "rex.yaml")); err == nil {
		m := parseManifest(stackDir)
		for _, c := range m.Spec.Containers {
			if c.Type == "compose" {
				contDir := filepath.Join(stackDir, c.Name)
				if _, err := os.Stat(contDir); err == nil {
					cmd := exec.Command("docker", "compose", "--project-name", target, "stop")
					cmd.Dir = contDir
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					cmd.Run()
				}
			}
		}
	}

	// Stop any raw image containers or matching containers
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

// CmdDown stops and removes containers, networks, and internal volumes (similar to docker compose down).
func CmdDown(args []string) {
	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}
	target, stackDir := ResolveWorkspaceOrDie(arg)

	logging.LogInfo("Tearing down stack '%s'...", target)

	if _, err := os.Stat(filepath.Join(stackDir, "rex.yaml")); err == nil {
		m := parseManifest(stackDir)
		for _, c := range m.Spec.Containers {
			if c.Type == "compose" {
				contDir := filepath.Join(stackDir, c.Name)
				if _, err := os.Stat(contDir); err == nil {
					cmd := exec.Command("docker", "compose", "--project-name", target, "down")
					cmd.Dir = contDir
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					cmd.Run()
				}
			} else if c.Type == "image" {
				containerName := fmt.Sprintf("%s-%s", c.Name, target)
				exec.Command("docker", "stop", containerName).Run()
				exec.Command("docker", "rm", containerName).Run()
			}
		}
	}

	// Stop and remove any remaining containers for this workspace
	out, err := exec.Command("docker", "ps", "-a", "-q", "--filter", fmt.Sprintf("label=com.docker.compose.project=%s", target)).Output()
	if err == nil {
		containerIDs := strings.Fields(string(out))
		if len(containerIDs) > 0 {
			stopArgs := append([]string{"stop"}, containerIDs...)
			exec.Command("docker", stopArgs...).Run()
			rmArgs := append([]string{"rm", "-f"}, containerIDs...)
			rmCmd := exec.Command("docker", rmArgs...)
			rmCmd.Stdout = os.Stdout
			rmCmd.Stderr = os.Stderr
			rmCmd.Run()
		}
	}

	logging.LogInfo("Stack '%s' brought down.", target)
}

func CmdSwitch(args []string) {
	if len(args) < 1 {
		Die("Usage: rexctl switch <stack>")
	}
	target, stackDir := ResolveWorkspaceOrDie(args[0])
	ValidateStack(stackDir)

	if active, found := getActiveStack(); found {
		if active == target {
			logging.LogInfo("Stack '%s' is already running.", target)
			return
		}
		logging.LogInfo("Stopping active workspace '%s' (preserving container states)...", active)
		CmdStop([]string{active})
	}

	logging.LogInfo("Switching to workspace '%s'...", target)
	CmdUp([]string{target})
}

func CmdDestroy(args []string) {
	if len(args) < 1 {
		Die("Usage: rexctl destroy <stack>")
	}
	target, stackDir := ResolveWorkspaceOrDie(args[0])

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

	// Always ensure containers/volumes are brought down before removing files
	CmdDown([]string{target})

	logging.LogInfo("Destroying stack '%s'...", target)

	if err := os.RemoveAll(stackDir); err != nil {
		Die("Failed to remove stack directory: %v", err)
	}

	logging.LogInfo("Stack '%s' destroyed.", target)
}

func CmdEdit(args []string) {
	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}
	stackName, stackDir := ResolveWorkspaceOrDie(arg)
	targetFile := filepath.Join(stackDir, "rex.yaml")

	if _, err := os.Stat(targetFile); err != nil {
		Die("Stack manifest not found: %s\nRun 'rexctl create %s' first.", targetFile, stackName)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	editorParts := strings.Fields(editor)
	cmdName := editorParts[0]
	cmdArgs := append(editorParts[1:], targetFile)

	cmd := exec.Command(cmdName, cmdArgs...)
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

var Die = func(format string, a ...any) {
	logging.LogErr(format, a...)
	os.Exit(1)
}

