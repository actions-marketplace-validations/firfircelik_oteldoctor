package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCmd_Defaults(t *testing.T) {
	cmd := newRootCmd()

	if cmd.Use != "oteldoctor" {
		t.Errorf("expected Use 'oteldoctor', got '%s'", cmd.Use)
	}

	if len(cmd.Commands()) != 5 {
		t.Errorf("expected 3 subcommands, got %d", len(cmd.Commands()))
	}

	names := make(map[string]bool)
	for _, c := range cmd.Commands() {
		names[c.Use] = true
	}

	if !names["analyze <path>"] {
		t.Error("expected 'analyze' subcommand")
	}

	if !names["version"] {
		t.Error("expected 'version' subcommand")
	}
}

func TestRootCmd_Help(t *testing.T) {
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "oteldoctor") {
		t.Error("help output should contain 'oteldoctor'")
	}
	if !strings.Contains(out, "analyze") {
		t.Error("help output should contain 'analyze'")
	}
	if !strings.Contains(out, "version") {
		t.Error("help output should contain 'version'")
	}
}
