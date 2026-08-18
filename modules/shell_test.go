package modules

import (
	"reflect"
	"testing"
)

func TestParseShellArgs_ContainerOnly(t *testing.T) {
	opts, err := ParseShellArgs([]string{"raptor_ws"})
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if opts.ContainerName != "raptor_ws" {
		t.Errorf("expected containerName 'raptor_ws', got '%s'", opts.ContainerName)
	}
}

func TestParseShellArgs_StackAndContainer(t *testing.T) {
	opts, err := ParseShellArgs([]string{"my-stack", "api_service"})
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if opts.StackName != "my-stack" {
		t.Errorf("expected stackName 'my-stack', got '%s'", opts.StackName)
	}
	if opts.ContainerName != "api_service" {
		t.Errorf("expected containerName 'api_service', got '%s'", opts.ContainerName)
	}
}

func TestParseShellArgs_CustomUserShortFlag(t *testing.T) {
	opts, err := ParseShellArgs([]string{"-u", "rex", "my-stack", "raptor_ws"})
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if opts.User != "rex" {
		t.Errorf("expected user 'rex', got '%s'", opts.User)
	}
	if opts.StackName != "my-stack" {
		t.Errorf("expected stackName 'my-stack', got '%s'", opts.StackName)
	}
	if opts.ContainerName != "raptor_ws" {
		t.Errorf("expected containerName 'raptor_ws', got '%s'", opts.ContainerName)
	}
}

func TestParseShellArgs_CustomUserLongFlag(t *testing.T) {
	opts, err := ParseShellArgs([]string{"--user", "rex", "raptor_ws"})
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if opts.User != "rex" {
		t.Errorf("expected user 'rex', got '%s'", opts.User)
	}
	if opts.ContainerName != "raptor_ws" {
		t.Errorf("expected containerName 'raptor_ws', got '%s'", opts.ContainerName)
	}
}

func TestParseShellArgs_MissingFlagValue(t *testing.T) {
	_, err := ParseShellArgs([]string{"-u"})
	if err == nil {
		t.Fatal("expected error for trailing -u without value, got nil")
	}
}

func TestParseShellArgs_NoArgs(t *testing.T) {
	_, err := ParseShellArgs([]string{})
	if err == nil {
		t.Fatal("expected error when no arguments provided, got nil")
	}
}

func TestBuildShellExecArgs_Defaults(t *testing.T) {
	args := BuildShellExecArgs("web-container", "root", "/bin/bash")
	expected := []string{"exec", "-it", "-u", "root", "web-container", "/bin/bash"}
	if !reflect.DeepEqual(args, expected) {
		t.Errorf("expected %v, got %v", expected, args)
	}
}

func TestBuildShellExecArgs_CustomUserAndShell(t *testing.T) {
	args := BuildShellExecArgs("robot-core", "rex", "/bin/sh")
	expected := []string{"exec", "-it", "-u", "rex", "robot-core", "/bin/sh"}
	if !reflect.DeepEqual(args, expected) {
		t.Errorf("expected %v, got %v", expected, args)
	}
}
