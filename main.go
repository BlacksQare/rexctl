package main

import (
	"os"

	"rexctl/logging"
	"rexctl/modules"
)

func main() {
	args := os.Args[1:]
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
	case "start":
		modules.CmdStart(cmdArgs)
	case "down", "stop":
		modules.CmdDown(cmdArgs)
	case "switch":
		modules.CmdDown(nil)
		modules.CmdStart(cmdArgs)
	case "destroy":
		modules.CmdDestroy(cmdArgs)
	case "sync":
		modules.CmdSync(cmdArgs)
	case "validate":
		cwd, _ := os.Getwd()
		modules.ValidateStack(cwd)
		logging.LogInfo("Stack configuration is valid.")
	case "edit":
		modules.CmdEdit(cmdArgs)
	case "info":
		modules.CmdInfo()
	case "help":
		modules.CmdHelp()
	default:
		modules.CmdHelp()
	}
}
