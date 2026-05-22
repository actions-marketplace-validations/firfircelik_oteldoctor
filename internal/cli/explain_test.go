package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestExplain_KnownRule(t *testing.T) {
	cmd := newExplainCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{"OTEL-REL-102"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()

	checks := []string{
		"Rule: OTEL-REL-102",
		"Title:",
		"Category: reliability",
		"Default Severity:",
		"Why it matters:",
		"Bad example:",
		"Good example:",
		"How to fix:",
		"memory_limiter",
	}
	for _, c := range checks {
		if !strings.Contains(out, c) {
			t.Errorf("expected output to contain %q", c)
		}
	}
}

func TestExplain_StructuralRule(t *testing.T) {
	cmd := newExplainCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	cmd.SetArgs([]string{"OTEL-STRUCT-001"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(strings.ToLower(out), "undefined") {
		t.Error("expected receiver reference docs")
	}
	if !strings.Contains(out, "Category: structural") {
		t.Error("expected structural category")
	}
}

func TestExplain_SecurityRule(t *testing.T) {
	cmd := newExplainCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	cmd.SetArgs([]string{"OTEL-SEC-202"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "secret") {
		t.Error("expected secret content")
	}
	if !strings.Contains(out, "Category: security") {
		t.Error("expected security category")
	}
}

func TestExplain_K8sRule(t *testing.T) {
	cmd := newExplainCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	cmd.SetArgs([]string{"OTEL-K8S-501"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Category: kubernetes") {
		t.Error("expected kubernetes category")
	}
}

func TestExplain_UnknownRule(t *testing.T) {
	cmd := newExplainCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{"OTEL-FAKE-999"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown rule")
	}
	if !strings.Contains(err.Error(), "unknown rule") {
		t.Errorf("expected unknown rule error, got: %v", err)
	}
}

func TestExplain_NoArgs(t *testing.T) {
	cmd := newExplainCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing argument")
	}
}
