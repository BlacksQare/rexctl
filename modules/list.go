package modules

import (
	"fmt"
	"os"
	"path/filepath"
	"rexctl/config"
)

func CmdList() {
	entries, err := os.ReadDir(config.WorkspacesDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			if _, err := os.Stat(filepath.Join(config.WorkspacesDir, e.Name(), "rex.yaml")); err == nil {
				fmt.Println(e.Name())
			}
		}
	}
}
