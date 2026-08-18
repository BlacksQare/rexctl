package modules

import (
	"fmt"
	"os"
	"path/filepath"
	"rexctl/structs"
)

// GetWorkspaceInfo gathers detailed information about containers in a workspace.
func GetWorkspaceInfo(stackDir string, stackName string, activeStack string) (structs.WorkspaceInfo, error) {
	m, err := ParseManifest(stackDir)
	if err != nil {
		return structs.WorkspaceInfo{}, err
	}

	isRunning := (activeStack != "" && activeStack == stackName)

	info := structs.WorkspaceInfo{
		StackName: stackName,
		IsRunning: isRunning,
	}

	for _, r := range m.Spec.Containers {
		cinfo := structs.ContainerInfo{
			Name:   r.Name,
			Type:   r.Type,
			Remote: r.Remote,
			State:  "N/A",
		}

		if r.Type == "compose" {
			cinfo.RequestedRevision = r.Revision
			repoDir := filepath.Join(stackDir, r.Name)
			cinfo.CurrentRevision = "MISSING"

			if _, err := os.Stat(repoDir); err == nil {
				rev, _ := GetGitCommitHash(repoDir)
				if rev != "" {
					cinfo.CurrentRevision = rev
				}
				dirty, _ := IsGitDirty(repoDir)
				if dirty {
					cinfo.State = "dirty"
				} else {
					cinfo.State = "clean"
				}
				cinfo.ImageTag = BuildImageTag(stackName, r.Name, cinfo.CurrentRevision, dirty)
			}
		} else if r.Type == "image" {
			cinfo.State = "remote image"
			cinfo.ImageTag = fmt.Sprintf("rexctl/%s/%s:latest", stackName, r.Name)
		}


		info.Containers = append(info.Containers, cinfo)
	}

	return info, nil
}

// CmdInfo displays detailed metadata about a workspace (from argument or current directory).
func CmdInfo(args []string) {
	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}
	stackName, stackDir := ResolveWorkspaceOrDie(arg)

	active, _ := GetActiveStack()
	info, err := GetWorkspaceInfo(stackDir, stackName, active)
	if err != nil {
		Die("Failed to get workspace info: %v", err)
	}

	runningStr := "no"
	if info.IsRunning {
		runningStr = "yes"
	}

	fmt.Printf("Stack: %s\nRunning: %s\n------------------------------------------------------------\n", info.StackName, runningStr)
	for _, c := range info.Containers {
		fmt.Printf("Repository:         %s\n", c.Name)
		fmt.Printf("Type:               %s\n", c.Type)
		fmt.Printf("Remote:             %s\n", c.Remote)

		if c.Type == "compose" {
			fmt.Printf("Requested Revision: %s\n", c.RequestedRevision)
			fmt.Printf("Current Revision:   %s\n", c.CurrentRevision)
		}

		if c.ImageTag != "" {
			fmt.Printf("Image Tag:          %s\n", c.ImageTag)
		}

		fmt.Printf("State:              %s\n", c.State)
		fmt.Println("------------------------------------------------------------")
	}
}
