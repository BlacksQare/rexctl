package main

import (
	"bytes"
	"io"
	"os"
	"rexctl/modules"
	"strings"
	"testing"
)

func captureStdout(f func()) string {
	r, w, _ := os.Pipe()
	stdout := os.Stdout
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = stdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestRun_NoArgsShowsHelp(t *testing.T) {
	out := captureStdout(func() {
		Run([]string{})
	})

	if !strings.Contains(out, "rexctl - Declarative Workspace Orchestrator") {
		t.Errorf("expected help output when no args passed, got: %s", out)
	}
}

func TestRun_HelpCommand(t *testing.T) {
	out := captureStdout(func() {
		Run([]string{"help"})
	})

	if !strings.Contains(out, "rexctl - Declarative Workspace Orchestrator") {
		t.Errorf("expected help output, got: %s", out)
	}
}

func TestRun_UnknownCommandShowsHelp(t *testing.T) {
	out := captureStdout(func() {
		Run([]string{"nonexistent-command-xyz"})
	})

	if !strings.Contains(out, "rexctl - Declarative Workspace Orchestrator") {
		t.Errorf("expected help output for unknown command, got: %s", out)
	}
}

func TestRun_ListCommand(t *testing.T) {
	Run([]string{"list"})
}

func TestRun_CommandDispatching(t *testing.T) {
	origDie := modules.Die
	modules.Die = func(format string, a ...any) {
		panic("die_called")
	}
	defer func() {
		modules.Die = origDie
	}()

	commands := []struct {
		name string
		args []string
	}{
		{"prepare-env", []string{"prepare-env", "nonexistent-ws"}},
		{"shell", []string{"shell"}},
		{"sh", []string{"sh"}},
		{"build", []string{"build"}},
		{"up", []string{"up"}},
		{"start", []string{"start"}},
		{"switch", []string{"switch"}},
		{"destroy", []string{"destroy"}},
		{"edit", []string{"edit"}},
		{"create", []string{"create"}},
		{"sync", []string{"sync"}},
		{"validate", []string{"validate"}},
		{"info", []string{"info", "nonexistent-ws"}},
		{"status", []string{"status", "nonexistent-ws"}},
		{"pwd", []string{"pwd"}},
		{"get", []string{"get"}},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				// We expect commands to execute dispatching (and potentially call Die or succeed)
				recover()
			}()
			Run(tc.args)
		})
	}
}
