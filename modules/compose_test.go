package modules

import (
	"reflect"
	"testing"
)

func TestBuildComposeArgs_Up(t *testing.T) {
	args := BuildComposeArgs("my-stack", "up", "-d")
	expected := []string{"compose", "--project-name", "my-stack", "up", "-d"}
	if !reflect.DeepEqual(args, expected) {
		t.Errorf("expected %v, got %v", expected, args)
	}
}

func TestBuildComposeArgs_Start(t *testing.T) {
	args := BuildComposeArgs("my-stack", "start")
	expected := []string{"compose", "--project-name", "my-stack", "start"}
	if !reflect.DeepEqual(args, expected) {
		t.Errorf("expected %v, got %v", expected, args)
	}
}

func TestBuildComposeArgs_Stop(t *testing.T) {
	args := BuildComposeArgs("my-stack", "stop")
	expected := []string{"compose", "--project-name", "my-stack", "stop"}
	if !reflect.DeepEqual(args, expected) {
		t.Errorf("expected %v, got %v", expected, args)
	}
}

func TestBuildComposeArgs_Down(t *testing.T) {
	args := BuildComposeArgs("my-stack", "down")
	expected := []string{"compose", "--project-name", "my-stack", "down"}
	if !reflect.DeepEqual(args, expected) {
		t.Errorf("expected %v, got %v", expected, args)
	}
}

func TestBuildComposeArgs_NoProjectName(t *testing.T) {
	args := BuildComposeArgs("", "ps")
	expected := []string{"compose", "ps"}
	if !reflect.DeepEqual(args, expected) {
		t.Errorf("expected %v, got %v", expected, args)
	}
}
