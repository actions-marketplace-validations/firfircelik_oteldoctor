package scanner

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestScan_FindsYAMLFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.yaml"), []byte("receivers:\n  otlp: {}\nexporters:\n  debug: {}\nservice:\n  pipelines:\n    traces:\n      receivers: [otlp]\n      exporters: [debug]"), 0644)
	os.WriteFile(filepath.Join(dir, "b.yml"), []byte("x: 1"), 0644)
	os.WriteFile(filepath.Join(dir, "c.txt"), []byte("hello"), 0644)

	s := New()
	files, err := s.Scan(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sort.Strings(files)
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d: %v", len(files), files)
	}
	if filepath.Base(files[0]) != "a.yaml" {
		t.Errorf("expected a.yaml, got %s", files[0])
	}
	if filepath.Base(files[1]) != "b.yml" {
		t.Errorf("expected b.yml, got %s", files[1])
	}
}

func TestScan_Recursive(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "sub"), 0755)
	os.WriteFile(filepath.Join(dir, "root.yaml"), []byte("x: 1"), 0644)
	os.WriteFile(filepath.Join(dir, "sub", "nested.yaml"), []byte("y: 1"), 0644)

	s := New()
	files, err := s.Scan(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
}

func TestScan_NonRecursive(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "sub"), 0755)
	os.WriteFile(filepath.Join(dir, "root.yaml"), []byte("x: 1"), 0644)
	os.WriteFile(filepath.Join(dir, "sub", "nested.yaml"), []byte("y: 1"), 0644)

	s := New()
	s.Recursive = false
	files, err := s.Scan(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
}

func TestScan_SkipsHiddenDirs(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".hidden"), 0755)
	os.WriteFile(filepath.Join(dir, ".hidden", "secret.yaml"), []byte("x: 1"), 0644)
	os.WriteFile(filepath.Join(dir, "visible.yaml"), []byte("y: 1"), 0644)

	s := New()
	files, err := s.Scan(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 visible file, got %d: %v", len(files), files)
	}
}

func TestScan_SkipsNodeModules(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "node_modules"), 0755)
	os.WriteFile(filepath.Join(dir, "node_modules", "config.yaml"), []byte("x: 1"), 0644)
	os.WriteFile(filepath.Join(dir, "app.yaml"), []byte("y: 1"), 0644)

	s := New()
	files, err := s.Scan(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d: %v", len(files), files)
	}
}

func TestIsCollectorConfig_True(t *testing.T) {
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

	ok, err := IsCollectorConfig(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected collector config detection")
	}
}

func TestIsCollectorConfig_False(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"empty", "{}"},
		{"random", "apiVersion: v1\nkind: Pod\n"},
		{"single key", "receivers:\n  otlp: {}\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			f := filepath.Join(dir, "cfg.yaml")
			os.WriteFile(f, []byte(tt.content), 0644)

			ok, err := IsCollectorConfig(f)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ok {
				t.Error("expected non-collector config")
			}
		})
	}
}

func TestIsCollectorConfig_ProcessorsAndExporters(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "cfg.yaml")
	os.WriteFile(f, []byte(`processors:
  batch: {}
exporters:
  debug: {}
`), 0644)

	ok, err := IsCollectorConfig(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected collector config (2 matching keys)")
	}
}

func TestFilterByGlob_Include(t *testing.T) {
	files := []string{"a.yaml", "b.yaml", "c.yml", "d.txt"}
	result := FilterByGlob(files, []string{"*.yaml"}, false)
	if len(result) != 2 {
		t.Errorf("expected 2 files matching *.yaml, got %d", len(result))
	}
}

func TestFilterByGlob_Exclude(t *testing.T) {
	files := []string{"prod.yaml", "dev.yaml", "test.yaml"}
	result := FilterByGlob(files, []string{"dev*"}, true)
	if len(result) != 2 {
		t.Errorf("expected 2 files after excluding dev*, got %d", len(result))
	}
	for _, f := range result {
		if filepath.Base(f) == "dev.yaml" {
			t.Error("dev.yaml should have been excluded")
		}
	}
}

func TestFilterCollectorConfigs(t *testing.T) {
	dir := t.TempDir()
	f1 := filepath.Join(dir, "collector.yaml")
	f2 := filepath.Join(dir, "deployment.yaml")
	f3 := filepath.Join(dir, "random.yaml")
	os.WriteFile(f1, []byte("receivers:\n  otlp: {}\nexporters:\n  debug: {}\n"), 0644)
	os.WriteFile(f2, []byte("apiVersion: apps/v1\nkind: Deployment\n"), 0644)
	os.WriteFile(f3, []byte("processors:\n  batch: {}\nexporters:\n  debug: {}\n"), 0644)

	result, err := FilterCollectorConfigs([]string{f1, f2, f3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 collector configs, got %d: %v", len(result), result)
	}
}

func TestFilterCollectorConfigs_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "bad.yaml")
	os.WriteFile(f, []byte("\tinvalid: [[["), 0644)

	result, err := FilterCollectorConfigs([]string{f})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 collector configs for invalid YAML, got %d", len(result))
	}
}
