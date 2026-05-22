package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGraphCmd_Mermaid(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "cfg.yaml")
	os.WriteFile(f, []byte(`receivers:
  otlp:
    protocols:
      grpc: {}
processors:
  memory_limiter:
    limit_mib: 512
  batch:
    timeout: 200ms
exporters:
  debug:
    verbosity: detailed
service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [memory_limiter, batch]
      exporters: [debug]
`), 0644)

	cmd := newGraphCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	cmd.SetArgs([]string{f})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()

	if !strings.Contains(out, "flowchart LR") {
		t.Error("expected mermaid flowchart header")
	}
	if !strings.Contains(out, "subgraph traces") {
		t.Error("expected traces subgraph")
	}
	if !strings.Contains(out, "otlp") {
		t.Error("expected otlp receiver")
	}
	if !strings.Contains(out, "memory_limiter") {
		t.Error("expected memory_limiter processor")
	}
	if !strings.Contains(out, "batch") {
		t.Error("expected batch processor")
	}
	if !strings.Contains(out, "debug") {
		t.Error("expected debug exporter")
	}
	if !strings.Contains(out, "-->") {
		t.Error("expected edges between nodes")
	}
}

func TestGraphCmd_Mermaid_MultiplePipelines(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "cfg.yaml")
	os.WriteFile(f, []byte(`receivers:
  otlp:
    protocols:
      grpc: {}
processors:
  batch:
    timeout: 200ms
exporters:
  debug:
    verbosity: detailed
service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [debug]
    metrics:
      receivers: [otlp]
      processors: []
      exporters: [debug]
    logs:
      receivers: [otlp]
      processors: [batch]
      exporters: [debug]
`), 0644)

	cmd := newGraphCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	cmd.SetArgs([]string{f})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()

	if !strings.Contains(out, "subgraph traces") {
		t.Error("expected traces subgraph")
	}
	if !strings.Contains(out, "subgraph metrics") {
		t.Error("expected metrics subgraph")
	}
	if !strings.Contains(out, "subgraph logs") {
		t.Error("expected logs subgraph")
	}
}

func TestGraphCmd_Mermaid_NamedComponents(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "cfg.yaml")
	os.WriteFile(f, []byte(`receivers:
  otlp/datadog:
    protocols:
      grpc: {}
processors:
  batch/default:
    timeout: 200ms
exporters:
  debug/inspect:
    verbosity: detailed
service:
  pipelines:
    traces:
      receivers: [otlp/datadog]
      processors: [batch/default]
      exporters: [debug/inspect]
`), 0644)

	cmd := newGraphCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	cmd.SetArgs([]string{f})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()

	// Named components should appear in the graph with their full IDs
	if !strings.Contains(out, "otlp/datadog") {
		t.Error("expected otlp/datadog")
	}
	if !strings.Contains(out, "batch/default") {
		t.Error("expected batch/default")
	}
	if !strings.Contains(out, "debug/inspect") {
		t.Error("expected debug/inspect")
	}
}

func TestGraphCmd_DOT(t *testing.T) {
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

	cmd := newGraphCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	cmd.SetArgs([]string{"--format", "dot", f})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()

	if !strings.Contains(out, "digraph oteldoctor") {
		t.Error("expected digraph header")
	}
	if !strings.Contains(out, "rankdir=LR") {
		t.Error("expected rankdir")
	}
	if !strings.Contains(out, "->") {
		t.Error("expected edges")
	}
	if !strings.Contains(out, "subgraph cluster_traces") {
		t.Error("expected traces cluster")
	}
}

func TestGraphCmd_JSON(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "cfg.yaml")
	os.WriteFile(f, []byte(`receivers:
  otlp:
    protocols:
      grpc: {}
processors:
  batch:
    timeout: 200ms
exporters:
  debug:
    verbosity: detailed
service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [debug]
`), 0644)

	cmd := newGraphCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	cmd.SetArgs([]string{"--format", "json", f})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()

	var jg jsonGraph
	if err := json.Unmarshal([]byte(out), &jg); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, out)
	}

	if len(jg.Pipelines) != 1 {
		t.Fatalf("expected 1 pipeline, got %d", len(jg.Pipelines))
	}

	pl := jg.Pipelines[0]
	if pl.Signal != "traces" {
		t.Errorf("expected signal traces, got %q", pl.Signal)
	}
	if len(pl.Receivers) != 1 || pl.Receivers[0] != "otlp" {
		t.Errorf("expected receivers [otlp], got %v", pl.Receivers)
	}
	if len(pl.Processors) != 1 || pl.Processors[0] != "batch" {
		t.Errorf("expected processors [batch], got %v", pl.Processors)
	}
	if len(pl.Exporters) != 1 || pl.Exporters[0] != "debug" {
		t.Errorf("expected exporters [debug], got %v", pl.Exporters)
	}

	if len(jg.Nodes) < 3 {
		t.Errorf("expected at least 3 nodes, got %d", len(jg.Nodes))
	}
	if len(jg.Edges) < 2 {
		t.Errorf("expected at least 2 edges, got %d", len(jg.Edges))
	}
}

func TestGraphCmd_ParseError(t *testing.T) {
	cmd := newGraphCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{"/nonexistent/path.yaml"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}
