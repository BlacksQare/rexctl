package modules

import (
	"fmt"
	"os"
	"path/filepath"
	"rexctl/config"
)

// ListWorkspaces scans a directory for workspace directories containing a valid rex.yaml.
func ListWorkspaces(workspacesDir string) ([]string, error) {
	entries, err := os.ReadDir(workspacesDir)
	if err != nil {
		return nil, err
	}

	var results []string
	for _, e := range entries {
		if e.IsDir() {
			if _, err := os.Stat(filepath.Join(workspacesDir, e.Name(), "rex.yaml")); err == nil {
				results = append(results, e.Name())
			}
		}
	}
	return results, nil
}

func CmdList() {
	workspaces, err := ListWorkspaces(config.WorkspacesDir)
	if err != nil {
		return
	}
	for _, ws := range workspaces {
		fmt.Println(ws)
	}
}

