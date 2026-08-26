package modules

import (
	"rexctl/config"
	"strings"
	"testing"
)

func TestGetBuildCommitHash_Injected(t *testing.T) {
	orig := config.CommitHash
	config.CommitHash = "abcdef1"
	defer func() { config.CommitHash = orig }()

	hash := GetBuildCommitHash()
	if hash != "abcdef1" {
		t.Errorf("expected 'abcdef1', got '%s'", hash)
	}
}

func TestCmdVersion_Output(t *testing.T) {
	orig := config.CommitHash
	config.CommitHash = "1234567"
	defer func() { config.CommitHash = orig }()

	out := captureStdout(func() {
		CmdVersion()
	})

	trimmed := strings.TrimSpace(out)
	if trimmed != "1234567" {
		t.Errorf("expected '1234567', got '%s'", trimmed)
	}
}
