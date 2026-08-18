package modules

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"rexctl/config"
	"rexctl/structs"
	"strings"

	"gopkg.in/yaml.v3"
)

// ResolveWorkspace determines the target stack name and directory on disk.
// If arg is given: checks config.WorkspacesDir/<arg> and direct directory <arg>.
// If arg is empty: checks current working directory for rex.yaml.
func ResolveWorkspace(arg string) (string, string, error) {
	if arg != "" {
		wsPath := filepath.Join(config.WorkspacesDir, arg)
		if _, err := os.Stat(filepath.Join(wsPath, "rex.yaml")); err == nil {
			return arg, wsPath, nil
		}
		if _, err := os.Stat(filepath.Join(arg, "rex.yaml")); err == nil {
			absPath, err := filepath.Abs(arg)
			if err == nil {
				return filepath.Base(absPath), absPath, nil
			}
			return filepath.Base(arg), arg, nil
		}
		return arg, wsPath, nil
	}

	cwd, err := os.Getwd()
	if err == nil {
		if _, statErr := os.Stat(filepath.Join(cwd, "rex.yaml")); statErr == nil {
			return filepath.Base(cwd), cwd, nil
		}
	}

	return "", "", fmt.Errorf("no stack specified and no rex.yaml found in current directory")
}

// ResolveWorkspaceOrDie resolves the workspace or exits with error.
func ResolveWorkspaceOrDie(arg string) (string, string) {
	name, dir, err := ResolveWorkspace(arg)
	if err != nil {
		Die(err.Error())
	}
	return name, dir
}

// GetTargetStack returns the target stack name from argument or current directory rex.yaml.
func GetTargetStack(arg string) (string, error) {
	name, _, err := ResolveWorkspace(arg)
	return name, err
}

func getTargetStack(arg string) string {
	name, _ := ResolveWorkspaceOrDie(arg)
	return name
}

// ParseManifest reads and parses a rex.yaml manifest from stackDir.
func ParseManifest(stackDir string) (structs.Manifest, error) {
	targetFile := filepath.Join(stackDir, "rex.yaml")
	data, err := os.ReadFile(targetFile)
	if err != nil {
		return structs.Manifest{}, fmt.Errorf("manifest rex.yaml is missing or unreadable in %s: %w", stackDir, err)
	}

	var m structs.Manifest

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	if err := decoder.Decode(&m); err != nil {
		errMsg := err.Error()
		re := regexp.MustCompile(` in type structs\.[a-zA-Z]+`)
		cleanMsg := re.ReplaceAllString(errMsg, "")
		cleanMsg = strings.ReplaceAll(cleanMsg, "yaml: unmarshal errors:", "Unrecognized fields or formatting issues:")
		return structs.Manifest{}, fmt.Errorf("invalid YAML structure in rex.yaml:\n%s", cleanMsg)
	}

	return m, nil
}

func parseManifest(stackDir string) structs.Manifest {
	m, err := ParseManifest(stackDir)
	if err != nil {
		Die(err.Error())
	}
	return m
}

// ValidateStackDir validates the workspace manifest within stackDir.
func ValidateStackDir(stackDir string) error {
	m, err := ParseManifest(stackDir)
	if err != nil {
		return err
	}

	if m.Kind != "RexctlWorkspace" {
		return fmt.Errorf("validation failed: Manifest 'kind' must be 'RexctlWorkspace', got '%s'", m.Kind)
	}
	if len(m.Spec.Containers) == 0 {
		return fmt.Errorf("validation failed: Manifest 'spec.containers' is empty or missing")
	}

	containerNames := make(map[string]bool)
	for _, c := range m.Spec.Containers {
		if c.Name == "" || c.Remote == "" || c.Type == "" {
			return fmt.Errorf("validation failed: Container '%s' is missing required fields (name, type, remote)", c.Name)
		}

		if c.Type != "compose" && c.Type != "image" {
			return fmt.Errorf("validation failed: Container '%s' has unknown type '%s' (must be 'compose' or 'image')", c.Name, c.Type)
		}

		if c.Type == "compose" && c.Revision == "" {
			return fmt.Errorf("validation failed: Compose container '%s' is missing required field 'revision'", c.Name)
		}

		if containerNames[c.Name] {
			return fmt.Errorf("validation failed: Container name '%s' is duplicated", c.Name)
		}
		containerNames[c.Name] = true
	}

	return nil
}

func ValidateStack(stackDir string) {
	if err := ValidateStackDir(stackDir); err != nil {
		Die(err.Error())
	}
}
