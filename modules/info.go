package modules

import (
	"fmt"
	"os"
	"path/filepath"
)

func CmdInfo() {
	cwd, _ := os.Getwd()
	if _, err := os.Stat("rex.yaml"); err != nil {
		Die("info must be executed from a stack root (directory with rex.yaml).")
	}

	stackName := filepath.Base(cwd)
	active, found := getActiveStack()
	isRunning := "no"
	if found && active == stackName {
		isRunning = "yes"
	}

	fmt.Printf("Stack: %s\nRunning: %s\n------------------------------------------------------------\n", stackName, isRunning)

	m := parseManifest(cwd)
	for _, r := range m.Spec.Containers {
		currRev := "N/A"
		state := "N/A"

		if r.Type == "compose" {
			repoDir := filepath.Join(cwd, r.Name)
			currRev = "MISSING"

			if _, err := os.Stat(repoDir); err == nil {
				rev, _ := runCmd(repoDir, "git", "rev-parse", "HEAD")
				currRev = rev
				status, _ := runCmd(repoDir, "git", "status", "--porcelain")
				if status != "" {
					state = "dirty"
				} else {
					state = "clean"
				}
			}
		} else if r.Type == "image" {
			state = "remote image"
		}

		fmt.Printf("Repository:         %s\n", r.Name)
		fmt.Printf("Type:               %s\n", r.Type)
		fmt.Printf("Remote:             %s\n", r.Remote)

		if r.Type == "compose" {
			fmt.Printf("Requested Revision: %s\n", r.Revision)
			fmt.Printf("Current Revision:   %s\n", currRev)
		}

		fmt.Printf("State:              %s\n", state)
		fmt.Println("------------------------------------------------------------")
	}
}
