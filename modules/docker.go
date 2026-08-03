package modules

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/docker/docker/client"
	"gopkg.in/yaml.v3"
)

type overrideCompose struct {
	Services map[string]overrideService `yaml:"services"`
}

type overrideService struct {
	ContainerName string `yaml:"container_name,omitempty"`
}

func getDockerClient() *client.Client {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		Die("Failed to initialize Docker client: %v", err)
	}
	return cli
}

func getActiveStack() (string, bool) {
	cmd := exec.Command("docker", "ps",
		"--filter", "status=running",
		"--format", "{{.Label \"com.docker.compose.project\"}}")

	out, err := cmd.Output()
	if err != nil {
		return "", false
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			return line, true
		}
	}

	return "", false
}

func createComposeOverride(stackDir, relComposeFile, targetStack string) string {
	absComposeFile := filepath.Join(stackDir, relComposeFile)
	data, err := os.ReadFile(absComposeFile)
	if err != nil {
		return ""
	}

	var comp overrideCompose
	if err := yaml.Unmarshal(data, &comp); err != nil {
		return ""
	}

	override := overrideCompose{
		Services: make(map[string]overrideService),
	}

	hasOverrides := false
	for svcName, svc := range comp.Services {
		if svc.ContainerName != "" {
			override.Services[svcName] = overrideService{
				ContainerName: fmt.Sprintf("%s-%s", svc.ContainerName, targetStack),
			}
			hasOverrides = true
		}
	}

	if !hasOverrides {
		return ""
	}

	overrideData, err := yaml.Marshal(&override)
	if err != nil {
		return ""
	}

	overrideName := filepath.Base(relComposeFile) + ".rex.yaml"
	overridePath := filepath.Join(filepath.Dir(absComposeFile), overrideName)
	os.WriteFile(overridePath, overrideData, 0644)

	return filepath.Join(filepath.Dir(relComposeFile), overrideName)
}
