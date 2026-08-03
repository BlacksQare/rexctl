package modules

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"rexctl/structs"
	"strings"

	"gopkg.in/yaml.v3"
)

func getTargetStack(arg string) string {
	if arg != "" {
		return arg
	}
	if _, err := os.Stat("rex.yaml"); err == nil {
		cwd, _ := os.Getwd()
		return filepath.Base(cwd)
	}
	Die("No stack specified and no rex.yaml found in current directory.")
	return ""
}

func parseManifest(stackDir string) structs.Manifest {
	targetFile := filepath.Join(stackDir, "rex.yaml")
	data, err := os.ReadFile(targetFile)
	if err != nil {
		Die("Manifest rex.yaml is missing or unreadable in %s.", stackDir)
	}

	var m structs.Manifest

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	if err := decoder.Decode(&m); err != nil {
		errMsg := err.Error()
		re := regexp.MustCompile(` in type structs\.[a-zA-Z]+`)
		cleanMsg := re.ReplaceAllString(errMsg, "")
		cleanMsg = strings.ReplaceAll(cleanMsg, "yaml: unmarshal errors:", "Unrecognized fields or formatting issues:")
		Die("Invalid YAML structure in rex.yaml:\n%s", cleanMsg)
	}

	for i, c := range m.Spec.Containers {
		if c.Type == "compose" && c.ComposeFile == "" {
			m.Spec.Containers[i].ComposeFile = "docker-compose.yaml"
		}
	}

	return m
}

func ValidateStack(stackDir string) {
	m := parseManifest(stackDir)

	if m.Kind != "RexctlWorkspace" {
		Die("Validation failed: Manifest 'kind' must be 'RexctlWorkspace', got '%s'", m.Kind)
	}
	if len(m.Spec.Containers) == 0 {
		Die("Validation failed: Manifest 'spec.containers' is empty or missing.")
	}

	containerNames := make(map[string]bool)
	for _, c := range m.Spec.Containers {
		if c.Name == "" || c.Remote == "" || c.Type == "" {
			Die("Validation failed: Container '%s' is missing required fields (name, type, remote).", c.Name)
		}

		if c.Type != "compose" && c.Type != "image" {
			Die("Validation failed: Container '%s' has unknown type '%s' (must be 'compose' or 'image').", c.Name, c.Type)
		}

		if c.Type == "compose" && c.Revision == "" {
			Die("Validation failed: Compose container '%s' is missing required field 'revision'.", c.Name)
		}

		if containerNames[c.Name] {
			Die("Validation failed: Container name '%s' is duplicated.", c.Name)
		}
		containerNames[c.Name] = true

		if c.Type == "compose" {
			contDir := filepath.Join(stackDir, c.Name)
			if _, err := os.Stat(contDir); err == nil {
				ComposeFile := filepath.Join(contDir, c.ComposeFile)
				data, err := os.ReadFile(ComposeFile)
				if err != nil {
					Die("Validation failed: %s missing %s.", c.Name, c.ComposeFile)
				}

				var comp structs.ComposeFile
				if err := yaml.Unmarshal(data, &comp); err != nil {
					Die("Validation failed: Invalid YAML in %s/%s.", c.Name, c.ComposeFile)
				}
			}
		}
	}
}
