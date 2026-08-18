package modules

import (
	"fmt"
	"os"
	"os/exec"
	"rexctl/config"
	"rexctl/logging"
)

// ShellOptions represents parsed parameters for rexctl shell.
type ShellOptions struct {
	StackName     string
	ContainerName string
	User          string
}

// ParseShellArgs parses command line arguments for the shell command.
// Supported syntaxes:
//   rexctl shell [-u user] <container>
//   rexctl shell [-u user] <workspace> <container>
func ParseShellArgs(args []string) (ShellOptions, error) {
	opts := ShellOptions{
		User: config.DefaultShellUser,
	}

	var positional []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-u" || arg == "--user" {
			if i+1 < len(args) {
				opts.User = args[i+1]
				i++
			} else {
				return opts, fmt.Errorf("flag '%s' requires a user argument", arg)
			}
		} else {
			positional = append(positional, arg)
		}
	}

	if len(positional) == 0 {
		return opts, fmt.Errorf("missing container name. Usage: rexctl shell [-u user] [workspace] <container>")
	} else if len(positional) == 1 {
		active, found := GetActiveStack()
		if found && active != "" {
			opts.StackName = active
		}
		opts.ContainerName = positional[0]
	} else {
		opts.StackName = positional[0]
		opts.ContainerName = positional[1]
	}

	return opts, nil
}

// BuildShellExecArgs creates the docker exec arguments for launching an interactive shell.
func BuildShellExecArgs(containerIdentifier string, user string, shellCmd string) []string {
	if user == "" {
		user = config.DefaultShellUser
	}
	if shellCmd == "" {
		shellCmd = "/bin/bash"
	}
	return []string{"exec", "-it", "-u", user, containerIdentifier, shellCmd}
}

// CmdShell starts an interactive terminal shell inside a target container.
func CmdShell(args []string) {
	opts, err := ParseShellArgs(args)
	if err != nil {
		Die("%v", err)
	}

	targetContainer := ResolveRunningContainer(opts.StackName, opts.ContainerName)

	logging.LogInfo("Connecting to container '%s' as user '%s'...", targetContainer, opts.User)

	// Detect if /bin/bash is available in target container, fallback to /bin/sh
	shellCmd := "/bin/sh"
	testBash := exec.Command("docker", "exec", targetContainer, "test", "-x", "/bin/bash")
	if testBash.Run() == nil {
		shellCmd = "/bin/bash"
	}

	execArgs := BuildShellExecArgs(targetContainer, opts.User, shellCmd)
	cmd := exec.Command("docker", execArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		Die("Shell session ended or failed in container '%s': %v", targetContainer, err)
	}
}
