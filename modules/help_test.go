package modules

import (
	"strings"
	"testing"
)

func TestCmdHelp_ContainsAllCommands(t *testing.T) {
	out := captureStdout(func() {
		CmdHelp()
	})

	expectedCommands := []string{
		"create",
		"edit",
		"sync",
		"prepare-env",
		"build",
		"up",
		"start",
		"stop",
		"down",
		"shell / sh",
		"switch",
		"destroy",
		"get",
		"info",
		"status",
		"validate",
		"help",
	}

	for _, cmd := range expectedCommands {
		if !strings.Contains(out, cmd) {
			t.Errorf("expected help output to mention '%s'", cmd)
		}
	}
}
