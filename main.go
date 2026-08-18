package main

import (
	"os"

	"rexctl/logging"
	"rexctl/modules"
)

func Run(args []string) {
	if len(args) == 0 {
		modules.CmdHelp()
		return
	}

	command := args[0]
	cmdArgs := args[1:]

	switch command {
	case "create":
		modules.CmdCreate(cmdArgs)
	case "list":
		modules.CmdList()
	case "pwd":
		modules.CmdPwd(cmdArgs)
	case "get":
		modules.CmdGet()
	case "status":
		modules.CmdStatus()
	case "up":
		modules.CmdUp(cmdArgs)
	case "start":
		modules.CmdStart(cmdArgs)
	case "stop":
		modules.CmdStop(cmdArgs)
	case "down":
		modules.CmdDown(cmdArgs)

	case "prepare-env":
		modules.CmdPrepareEnv(cmdArgs)
	case "shell", "sh":
		modules.CmdShell(cmdArgs)
	case "switch":
		modules.CmdSwitch(cmdArgs)

	case "destroy":
		modules.CmdDestroy(cmdArgs)
	case "sync":
		modules.CmdSync(cmdArgs)
	case "validate":
		arg := ""
		if len(cmdArgs) > 0 {
			arg = cmdArgs[0]
		}
		_, stackDir := modules.ResolveWorkspaceOrDie(arg)
		modules.ValidateStack(stackDir)
		logging.LogInfo("Stack configuration is valid.")
	case "edit":
		modules.CmdEdit(cmdArgs)
	case "info":
		modules.CmdInfo(cmdArgs)
	case "help":
		modules.CmdHelp()
	default:
		modules.CmdHelp()
	}
}

func main() {
	Run(os.Args[1:])
}

