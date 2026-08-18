package structs

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestManifestSerialization(t *testing.T) {
	manifest := Manifest{
		Kind: "RexctlWorkspace",
		Spec: Spec{
			Containers: []Container{
				{
					Name:     "backend",
					Type:     "compose",
					Remote:   "git@github.com:org/backend.git",
					Revision: "main",
				},
				{
					Name:   "redis",
					Type:   "image",
					Remote: "redis:7-alpine",
				},
			},
		},
	}

	data, err := yaml.Marshal(&manifest)
	if err != nil {
		t.Fatalf("failed to marshal manifest: %v", err)
	}

	var parsed Manifest
	err = yaml.Unmarshal(data, &parsed)
	if err != nil {
		t.Fatalf("failed to unmarshal manifest: %v", err)
	}

	if parsed.Kind != "RexctlWorkspace" {
		t.Errorf("expected Kind 'RexctlWorkspace', got '%s'", parsed.Kind)
	}
	if len(parsed.Spec.Containers) != 2 {
		t.Fatalf("expected 2 containers, got %d", len(parsed.Spec.Containers))
	}
	if parsed.Spec.Containers[0].Name != "backend" || parsed.Spec.Containers[0].Revision != "main" {
		t.Errorf("unexpected container 0: %+v", parsed.Spec.Containers[0])
	}
	if parsed.Spec.Containers[1].Name != "redis" || parsed.Spec.Containers[1].Revision != "" {
		t.Errorf("unexpected container 1: %+v", parsed.Spec.Containers[1])
	}
}

func TestComposeFileParsing(t *testing.T) {
	composeYAML := `
services:
  web:
    container_name: custom-web
    image: nginx:latest
    build: .
  db:
    image: postgres:15
`
	var comp ComposeFile
	err := yaml.Unmarshal([]byte(composeYAML), &comp)
	if err != nil {
		t.Fatalf("failed to parse ComposeFile: %v", err)
	}

	if len(comp.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(comp.Services))
	}

	webSvc, ok := comp.Services["web"]
	if !ok {
		t.Fatal("expected 'web' service")
	}
	if webSvc.ContainerName != "custom-web" {
		t.Errorf("expected container_name 'custom-web', got '%s'", webSvc.ContainerName)
	}
	if webSvc.Image != "nginx:latest" {
		t.Errorf("expected image 'nginx:latest', got '%s'", webSvc.Image)
	}

	dbSvc, ok := comp.Services["db"]
	if !ok {
		t.Fatal("expected 'db' service")
	}
	if dbSvc.ContainerName != "" {
		t.Errorf("expected empty container_name for db, got '%s'", dbSvc.ContainerName)
	}
}
