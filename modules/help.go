package modules

import "fmt"

func CmdHelp() {
	fmt.Println(`rexctl - Declarative Workspace Orchestrator

Usage:
  rexctl <command> [arguments]

Commands:
  create       <workspace>           Initialize a new workspace with a default rex.yaml manifest
  edit         [workspace]           Open the workspace manifest in your default $EDITOR
  sync         [workspace]           Clone/pull repositories and prepare container states
  prepare-env  [workspace]           Execute environment init scripts across all repositories
  build        [workspace]           Build container images without starting them (docker compose build)
  up           <workspace>           Create and start the workspace containers (docker compose up -d)
  start        <workspace>           Start existing stopped workspace containers (docker compose start)
  stop         [workspace]           Stop running workspace containers without removing them (docker compose stop)
  down         [workspace]           Stop and remove workspace containers and networks (docker compose down)
  shell / sh   [-u user] [ws] <cont> Connect to a running container shell
  switch       <workspace>           Stop the current workspace and start a new one
  destroy      <workspace>           Safely remove a workspace directory completely from disk
  get                                Show the currently running workspace
  info         [workspace]           Display detailed information about a workspace
  status                             Show status of running containers
  validate                           Validate the workspace manifest
  version                            Print the build commit hash
  help                               Display this help message


Note: If [workspace] is omitted, rexctl will attempt to infer the workspace from your current directory.`)
}

