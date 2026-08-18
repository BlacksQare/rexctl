package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogInfo(t *testing.T) {
	var buf bytes.Buffer
	Stdout = &buf
	defer func() { Stdout = nil }()

	LogInfo("Test message %d", 42)
	output := buf.String()

	if !strings.Contains(output, "INFO") {
		t.Errorf("expected output to contain INFO, got: %s", output)
	}
	if !strings.Contains(output, "Test message 42") {
		t.Errorf("expected output to contain 'Test message 42', got: %s", output)
	}
}

func TestLogWarn(t *testing.T) {
	var buf bytes.Buffer
	Stderr = &buf
	defer func() { Stderr = nil }()

	LogWarn("Warning message %s", "caution")
	output := buf.String()

	if !strings.Contains(output, "WARN") {
		t.Errorf("expected output to contain WARN, got: %s", output)
	}
	if !strings.Contains(output, "Warning message caution") {
		t.Errorf("expected output to contain 'Warning message caution', got: %s", output)
	}
}

func TestLogErr(t *testing.T) {
	var buf bytes.Buffer
	Stderr = &buf
	defer func() { Stderr = nil }()

	LogErr("Error message %s", "critical")
	output := buf.String()

	if !strings.Contains(output, "ERROR") {
		t.Errorf("expected output to contain ERROR, got: %s", output)
	}
	if !strings.Contains(output, "Error message critical") {
		t.Errorf("expected output to contain 'Error message critical', got: %s", output)
	}
}
