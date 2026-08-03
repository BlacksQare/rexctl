package modules

import "fmt"

func CmdHelp() {
	fmt.Println(`rexctl - Declarative Workspace Orchestrator

Usage:
  rexctl <command> [arguments]

Commands:
  create  <workspace>   Initialize a new workspace with a default rex.yaml manifest
  edit    [workspace]   Open the workspace manifest in your default $EDITOR
  sync    [workspace]   Clone/pull repositories and prepare container states
  start   <workspace>   Start the specified workspace containers
  down                  Stop the currently running workspace (or 'stop')
  switch  <workspace>   Stop the current workspace and start a new one
  destroy <workspace>   Safely remove a workspace directory completely from disk
  get                   Show the currently running workspace
  info    [workspace]   Display detailed information about a workspace
  help                  Display this help message

Note: If [workspace] is omitted, rexctl will attempt to infer the workspace from your current directory.`)
}
