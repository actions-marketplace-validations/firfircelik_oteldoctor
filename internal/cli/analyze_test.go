package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
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

func TestAnalyzeCmd_FailOnFlag(t *testing.T) {
	cmd := newAnalyzeCmd()
	flag := cmd.Flags().Lookup("fail-on")
	if flag == nil {
		t.Fatal("expected 'fail-on' flag")
	}
	if flag.DefValue != "" {
		t.Errorf("expected default value '', got '%s'", flag.DefValue)
	}
}

func TestAnalyzeCmd_FileNotFound(t *testing.T) {
	cmd := newAnalyzeCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{"/nonexistent/path.yaml"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "reading config file") {
		t.Errorf("expected file read error, got: %v", err)
	}
}

func TestAnalyzeCmd_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "bad.yaml")
	os.WriteFile(f, []byte("receivers:\n\tinvalid:tab"), 0644)

	cmd := newAnalyzeCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{f})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "parse error") {
		t.Errorf("expected parse error, got: %v", err)
	}
}

func TestAnalyzeCmd_GoodConfig(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "good.yaml")
	os.WriteFile(f, []byte(`receivers:
  otlp:
    protocols:
      grpc:
        endpoint: "localhost:4317"
        tls:
          cert_file: /tls/cert.pem
processors:
  resource:
    attributes:
      - key: service.name
        value: my-service
        action: upsert
  memory_limiter:
    limit_mib: 512
  batch:
    timeout: 200ms
exporters:
  debug:
    verbosity: detailed
    retry_on_failure:
      enabled: true
    sending_queue:
      num_consumers: 10
    tls:
      insecure: false
service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [memory_limiter, resource, batch]
      exporters: [debug]
`), 0644)

	cmd := newAnalyzeCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{f})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "No issues found") {
		t.Errorf("expected 'No issues found', got: %s", out)
	}
}

func TestAnalyzeCmd_BadConfig_StructuralIssues(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "bad.yaml")
	os.WriteFile(f, []byte(`receivers:
  otlp:
    protocols:
      grpc:
        endpoint: "0.0.0.0:4317"
processors:
  batch: {}
exporters:
  debug: {}
service:
  pipelines:
    traces:
      receivers: [otlp, missing_rcv]
      processors: [batch, undefined_proc]
      exporters: [no_such_exporter]
`), 0644)

	cmd := newAnalyzeCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{f})
	err := cmd.Execute()

	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitError, got: %v", err)
	}
	if exitErr.Code != 1 {
		t.Errorf("expected exit code 1, got %d", exitErr.Code)
	}

	out := buf.String()
	if !strings.Contains(out, "OTEL-STRUCT-001") {
		t.Error("expected OTEL-STRUCT-001 (undefined receiver)")
	}
	if !strings.Contains(out, "OTEL-STRUCT-002") {
		t.Error("expected OTEL-STRUCT-002 (undefined processor)")
	}
	if !strings.Contains(out, "OTEL-STRUCT-003") {
		t.Error("expected OTEL-STRUCT-003 (undefined exporter)")
	}
}

func TestAnalyzeCmd_JSONFormat(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "good.yaml")
	os.WriteFile(f, []byte(`receivers:
  otlp:
    protocols:
      grpc:
        endpoint: "localhost:4317"
        tls:
          cert_file: /tls/cert.pem
processors:
  resource:
    attributes:
      - key: service.name
        value: my-service
        action: upsert
  memory_limiter:
    limit_mib: 512
  batch:
    timeout: 200ms
exporters:
  debug:
    verbosity: detailed
    retry_on_failure:
      enabled: true
    sending_queue:
      num_consumers: 10
    tls:
      insecure: false
service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [memory_limiter, resource, batch]
      exporters: [debug]
`), 0644)

	cmd := newAnalyzeCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{"--format", "json", f})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.HasPrefix(strings.TrimSpace(out), "[") {
		t.Errorf("expected JSON array output, got: %s", out)
	}
}

func TestAnalyzeCmd_FailOnThreshold(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "issues.yaml")
	os.WriteFile(f, []byte(`receivers:
  otlp:
    protocols:
      grpc: {}
processors:
  batch: {}
  orphan: {}
exporters:
  debug: {}
service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [debug]
`), 0644)

	// orphan processor is unused → severity low
	// With --fail-on medium, low issues should NOT trigger exit 1
	cmd := newAnalyzeCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{"--fail-on", "medium", f})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error with --fail-on medium, got: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "OTEL-STRUCT-006") {
		t.Error("expected OTEL-STRUCT-006 for orphan processor")
	}

	// With --fail-on low (default), should trigger exit 1
	cmd2 := newAnalyzeCmd()
	buf2 := new(bytes.Buffer)
	cmd2.SetOut(buf2)
	cmd2.SetErr(buf2)
	cmd2.SetArgs([]string{f})
	err2 := cmd2.Execute()

	var exitErr *ExitError
	if !errors.As(err2, &exitErr) {
		t.Fatalf("expected ExitError with default --fail-on low, got: %v", err2)
	}
	if exitErr.Code != 1 {
		t.Errorf("expected exit code 1, got %d", exitErr.Code)
	}
}

func TestAnalyzeCmd_InvalidFailOn(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "cfg.yaml")
	os.WriteFile(f, []byte(`receivers:
  otlp: {}
exporters:
  debug: {}
service:
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [debug]
`), 0644)

	cmd := newAnalyzeCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{"--fail-on", "nuclear", f})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid --fail-on")
	}
	if !strings.Contains(err.Error(), "invalid --fail-on") {
		t.Errorf("expected invalid fail-on error, got: %v", err)
	}
}
