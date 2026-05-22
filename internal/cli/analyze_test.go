package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestAnalyzeCmd_NoArgs(t *testing.T) {
	cmd := newAnalyzeCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing path argument")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestAnalyzeCmd_NotImplemented(t *testing.T) {
	cmd := newAnalyzeCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	cmd.SetArgs([]string{"test-config.yaml"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "not implemented") {
		t.Errorf("expected 'not implemented' in output, got: %s", out)
	}
}

func TestAnalyzeCmd_FormatFlag(t *testing.T) {
	cmd := newAnalyzeCmd()
	flag := cmd.Flags().Lookup("format")
	if flag == nil {
		t.Fatal("expected 'format' flag")
	}
	if flag.DefValue != "text" {
		t.Errorf("expected default value 'text', got '%s'", flag.DefValue)
	}
}
