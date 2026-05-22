package extractor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsConfigMap_True(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "cm.yaml")
	os.WriteFile(f, []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: otel-config
data:
  collector.yaml: |
    receivers:
      otlp:
        protocols:
          grpc: {}
    exporters:
      debug: {}
    service:
      pipelines:
        traces:
          receivers: [otlp]
          exporters: [debug]
`), 0644)

	ok, err := IsConfigMap(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected ConfigMap detection")
	}
}

func TestIsConfigMap_False(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"deployment", "apiVersion: apps/v1\nkind: Deployment\n"},
		{"service", "apiVersion: v1\nkind: Service\n"},
		{"no-data-key", "apiVersion: v1\nkind: ConfigMap\ndata:\n  other.yaml: hello\n"},
		{"empty", "{}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			f := filepath.Join(dir, "cm.yaml")
			os.WriteFile(f, []byte(tt.content), 0644)

			ok, err := IsConfigMap(f)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ok {
				t.Error("expected non-ConfigMap")
			}
		})
	}
}

func TestExtract_SingleConfig(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "cm.yaml")
	os.WriteFile(f, []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: otel-collector
data:
  collector.yaml: |
    receivers:
      otlp:
        protocols:
          grpc: {}
    exporters:
      debug: {}
    service:
      pipelines:
        traces:
          receivers: [otlp]
          exporters: [debug]
`), 0644)

	configs, err := Extract(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer Cleanup(configs)

	if len(configs) != 1 {
		t.Fatalf("expected 1 embedded config, got %d", len(configs))
	}

	c := configs[0]
	if c.ConfigMapName != "otel-collector" {
		t.Errorf("expected name otel-collector, got %q", c.ConfigMapName)
	}
	if c.DataKey != "collector.yaml" {
		t.Errorf("expected key collector.yaml, got %q", c.DataKey)
	}
	if c.SourceFile != f {
		t.Errorf("expected source file %s, got %s", f, c.SourceFile)
	}

	content, err := os.ReadFile(c.Path)
	if err != nil {
		t.Fatalf("reading temp file: %v", err)
	}

	if !strings.Contains(string(content), "receivers:") {
		t.Error("extracted content should contain receivers")
	}
	if !strings.Contains(string(content), "exporters:") {
		t.Error("extracted content should contain exporters")
	}
}

func TestExtract_MultipleKeys(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "cm.yaml")
	os.WriteFile(f, []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: multi-config
data:
  relay: |
    receivers:
      otlp: {}
    exporters:
      debug: {}
  config.yaml: |
    receivers:
      jaeger: {}
    exporters:
      otlp: {}
`), 0644)

	configs, err := Extract(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer Cleanup(configs)

	if len(configs) != 2 {
		t.Fatalf("expected 2 embedded configs, got %d", len(configs))
	}

	keys := map[string]bool{}
	for _, c := range configs {
		keys[c.DataKey] = true
	}
	if !keys["relay"] {
		t.Error("expected relay key")
	}
	if !keys["config.yaml"] {
		t.Error("expected config.yaml key")
	}
}

func TestExtract_NotConfigMap(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "deploy.yaml")
	os.WriteFile(f, []byte("apiVersion: apps/v1\nkind: Deployment\n"), 0644)

	configs, err := Extract(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(configs) != 0 {
		t.Errorf("expected 0 configs for non-ConfigMap, got %d", len(configs))
	}
}

func TestExtract_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "bad.yaml")
	os.WriteFile(f, []byte("\tinvalid: [[["), 0644)

	_, err := Extract(f)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestIsConfigMap_RelayKey(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "cm.yaml")
	os.WriteFile(f, []byte(`apiVersion: v1
kind: ConfigMap
data:
  relay: |
    receivers:
      otlp: {}
    exporters:
      debug: {}
`), 0644)

	ok, err := IsConfigMap(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected ConfigMap detection with relay key")
	}
}

func TestSourcePath(t *testing.T) {
	tests := []struct {
		file   string
		cmName string
		key    string
		want   string
	}{
		{"/path/to/cm.yaml", "otel", "collector.yaml", "cm.yaml/otel::collector.yaml"},
		{"/path/to/cm.yaml", "", "relay", "cm.yaml::relay"},
		{"cm.yaml", "my-config", "config.yaml", "cm.yaml/my-config::config.yaml"},
	}

	for _, tt := range tests {
		got := sourcePath(tt.file, tt.cmName, tt.key)
		if got != tt.want {
			t.Errorf("sourcePath(%q, %q, %q) = %q, want %q", tt.file, tt.cmName, tt.key, got, tt.want)
		}
	}
}
