package parser

import (
	"strings"
	"testing"

	"github.com/firfircelik/oteldoctor/internal/model"
	"gopkg.in/yaml.v3"
)

func readFixture(t *testing.T, path string) []byte {
	t.Helper()
	data, err := testDataFS.ReadFile(path)
	if err != nil {
		t.Fatalf("reading fixture %s: %v", path, err)
	}
	return data
}

func TestParser_Parse_MinimalConfig(t *testing.T) {
	p := New("testdata/valid_minimal.yaml")
	cfg, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Receivers) != 1 {
		t.Errorf("expected 1 receiver, got %d", len(cfg.Receivers))
	}
	if len(cfg.Processors) != 1 {
		t.Errorf("expected 1 processor, got %d", len(cfg.Processors))
	}
	if len(cfg.Exporters) != 1 {
		t.Errorf("expected 1 exporter, got %d", len(cfg.Exporters))
	}
	if len(cfg.Connectors) != 0 {
		t.Errorf("expected 0 connectors, got %d", len(cfg.Connectors))
	}
	if len(cfg.Extensions) != 0 {
		t.Errorf("expected 0 extensions, got %d", len(cfg.Extensions))
	}
	if len(cfg.Service.Pipelines) != 1 {
		t.Errorf("expected 1 pipeline, got %d", len(cfg.Service.Pipelines))
	}

	rcv, ok := cfg.Receivers["otlp"]
	if !ok {
		t.Fatal("expected 'otlp' receiver")
	}
	if rcv.ID != "otlp" {
		t.Errorf("expected ID 'otlp', got %q", rcv.ID)
	}
	if rcv.Kind != model.ComponentKindReceiver {
		t.Errorf("expected kind receiver, got %q", rcv.Kind)
	}
	if rcv.Name != "" {
		t.Errorf("expected empty name, got %q", rcv.Name)
	}
	if rcv.Location.File != "testdata/valid_minimal.yaml" {
		t.Errorf("expected file path, got %q", rcv.Location.File)
	}
	if rcv.Location.Line != 2 {
		t.Errorf("expected line 2, got %d", rcv.Location.Line)
	}

	pl, ok := cfg.Service.Pipelines["traces"]
	if !ok {
		t.Fatal("expected 'traces' pipeline")
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
}

func TestParser_Parse_NamedComponents(t *testing.T) {
	p := New("testdata/named_components.yaml")
	cfg, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		section string
		key     string
		wantID  string
		wantTyp string
	}{
		{"receivers", "otlp/datadog", "otlp/datadog", "otlp"},
		{"receivers", "otlp/newrelic", "otlp/newrelic", "otlp"},
		{"processors", "batch/default", "batch/default", "batch"},
		{"processors", "filter/errors", "filter/errors", "filter"},
		{"exporters", "debug/inspect", "debug/inspect", "debug"},
		{"exporters", "otlphttp/prod", "otlphttp/prod", "otlphttp"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			component, ok := getComponent(cfg, tt.section, tt.key)
			if !ok {
				t.Fatalf("component %q not found in %s", tt.key, tt.section)
			}
			if component.ID != tt.wantID {
				t.Errorf("expected ID %q, got %q", tt.wantID, component.ID)
			}
			if component.Name != strings.TrimPrefix(tt.key, tt.wantTyp+"/") {
				t.Errorf("unexpected name %q", component.Name)
			}
		})
	}

	if cfg.Service.Pipelines["metrics"].Receivers[0] != "otlp/datadog" {
		t.Errorf("expected otlp/datadog in metrics pipeline receivers")
	}

	if len(cfg.Service.Extensions) != 1 || cfg.Service.Extensions[0] != "health_check" {
		t.Errorf("expected service extensions [health_check], got %v", cfg.Service.Extensions)
	}
}

func TestParser_Parse_InvalidYAML(t *testing.T) {
	p := New("testdata/invalid.yaml")
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "expected a mapping") {
		t.Errorf("expected mapping error, got: %v", err)
	}
}

func TestParser_Parse_PipelineLineNumbers(t *testing.T) {
	p := New("testdata/pipelines.yaml")
	cfg, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		pipeline string
		wantLine int
	}{
		{"traces", 18},
		{"metrics", 22},
		{"logs", 26},
	}

	for _, tt := range tests {
		t.Run(tt.pipeline, func(t *testing.T) {
			pl, ok := cfg.Service.Pipelines[tt.pipeline]
			if !ok {
				t.Fatalf("pipeline %q not found", tt.pipeline)
			}
			if pl.Location.Line != tt.wantLine {
				t.Errorf("expected line %d, got %d", tt.wantLine, pl.Location.Line)
			}
			if pl.Location.File != "testdata/pipelines.yaml" {
				t.Errorf("expected file path, got %q", pl.Location.File)
			}
		})
	}
}

func TestParser_Parse_NoServiceSection(t *testing.T) {
	p := New("testdata/no_service.yaml")
	cfg, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Receivers) != 1 {
		t.Errorf("expected 1 receiver, got %d", len(cfg.Receivers))
	}
	if len(cfg.Processors) != 1 {
		t.Errorf("expected 1 processor, got %d", len(cfg.Processors))
	}
	if len(cfg.Exporters) != 1 {
		t.Errorf("expected 1 exporter, got %d", len(cfg.Exporters))
	}

	if len(cfg.Service.Pipelines) != 0 {
		t.Errorf("expected 0 pipelines, got %d", len(cfg.Service.Pipelines))
	}
	if len(cfg.Service.Extensions) != 0 {
		t.Errorf("expected 0 extensions, got %d", len(cfg.Service.Extensions))
	}
	if cfg.Service.Telemetry != nil {
		t.Errorf("expected nil telemetry")
	}
}

func TestParser_Parse_Connectors(t *testing.T) {
	p := New("testdata/pipelines.yaml")
	cfg, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Connectors) != 1 {
		t.Errorf("expected 1 connector, got %d", len(cfg.Connectors))
	}

	conn, ok := cfg.Connectors["count_connector"]
	if !ok {
		t.Fatal("expected 'count_connector' connector")
	}
	if conn.Kind != model.ComponentKindConnector {
		t.Errorf("expected kind connector, got %q", conn.Kind)
	}
	if conn.ID != "count_connector" {
		t.Errorf("expected ID 'count_connector', got %q", conn.ID)
	}

	metricsPL := cfg.Service.Pipelines["metrics"]
	hasConnector := false
	for _, r := range metricsPL.Receivers {
		if r == "count_connector" {
			hasConnector = true
		}
	}
	if !hasConnector {
		t.Error("expected count_connector in metrics receivers")
	}
}

func TestParser_Parse_Telemetry(t *testing.T) {
	p := New("testdata/pipelines.yaml")
	cfg, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Service.Telemetry == nil {
		t.Fatal("expected non-nil telemetry")
	}

	tm, ok := cfg.Service.Telemetry.(map[string]any)
	if !ok {
		t.Fatalf("expected telemetry to be map[string]any, got %T", cfg.Service.Telemetry)
	}
	if _, ok := tm["logs"]; !ok {
		t.Error("expected telemetry.logs")
	}
	if _, ok := tm["metrics"]; !ok {
		t.Error("expected telemetry.metrics")
	}
}

func TestParser_ParseBytes_EmptyDocument(t *testing.T) {
	p := New("dummy.yaml")
	_, err := p.ParseBytes([]byte{})
	if err == nil {
		t.Fatal("expected error for empty document")
	}
}

func TestParser_Parse_InvalidYAMLSyntax(t *testing.T) {
	p := New("testdata/syntax_error.yaml")
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected error for syntax error in YAML")
	}
	if !strings.Contains(err.Error(), "invalid YAML") {
		t.Errorf("expected invalid YAML error, got: %v", err)
	}
}

func TestParser_ParseBytes_NotMapping(t *testing.T) {
	p := New("dummy.yaml")
	_, err := p.ParseBytes([]byte("[1, 2, 3]"))
	if err == nil {
		t.Fatal("expected error for non-mapping root")
	}
	if !strings.Contains(err.Error(), "mapping") {
		t.Errorf("expected mapping error, got: %v", err)
	}
}

func TestParser_Parse_ConfigNodePreservesNestedLocation(t *testing.T) {
	p := New("testdata/valid_minimal.yaml")
	cfg, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rcv := cfg.Receivers["otlp"]
	if rcv.ConfigNode == nil {
		t.Fatal("expected non-nil ConfigNode")
	}
	if rcv.ConfigNode.Kind != yaml.MappingNode {
		t.Fatalf("expected ConfigNode to be a mapping, got kind %d", rcv.ConfigNode.Kind)
	}

	protocolsKey := findMappingKey(rcv.ConfigNode, "protocols")
	if protocolsKey == nil {
		t.Fatal("expected 'protocols' key in ConfigNode")
	}
	if protocolsKey.Line != 3 {
		t.Errorf("expected protocols key at line 3, got %d key line=%d", protocolsKey.Line, protocolsKey.Line)
	}

	protocolsVal := findMappingValue(rcv.ConfigNode, "protocols")
	if protocolsVal == nil {
		t.Fatal("expected 'protocols' value in ConfigNode")
	}
	grpcKey := findMappingKey(protocolsVal, "grpc")
	if grpcKey == nil {
		t.Fatal("expected 'grpc' key in protocols value")
	}
	if grpcKey.Line != 4 {
		t.Errorf("expected grpc key at line 4, got %d", grpcKey.Line)
	}

	grpcVal := findMappingValue(protocolsVal, "grpc")
	if grpcVal == nil {
		t.Fatal("expected 'grpc' value in protocols value")
	}
	endpointKey := findMappingKey(grpcVal, "endpoint")
	if endpointKey == nil {
		t.Fatal("expected 'endpoint' key in grpc value")
	}
	if endpointKey.Line != 5 {
		t.Errorf("expected endpoint key at line 5, got %d", endpointKey.Line)
	}
}

func TestParser_Parse_ConfigNodeNilForEmptyConfig(t *testing.T) {
	p := New("testdata/valid_minimal.yaml")
	cfg, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	batch := cfg.Processors["batch"]
	if batch.ConfigNode == nil {
		t.Fatal("expected non-nil ConfigNode for empty config")
	}
	if batch.ConfigNode.Kind != yaml.MappingNode {
		t.Fatalf("expected ConfigNode to be a mapping, got kind %d", batch.ConfigNode.Kind)
	}
}

func findMappingKey(node *yaml.Node, key string) *yaml.Node {
	if node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(node.Content)-1; i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i]
		}
	}
	return nil
}

func findMappingValue(node *yaml.Node, key string) *yaml.Node {
	if node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(node.Content)-1; i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

func getComponent(cfg *model.CollectorConfig, section, key string) (model.Component, bool) {
	var m map[string]model.Component
	switch section {
	case "receivers":
		m = cfg.Receivers
	case "processors":
		m = cfg.Processors
	case "exporters":
		m = cfg.Exporters
	case "connectors":
		m = cfg.Connectors
	case "extensions":
		m = cfg.Extensions
	default:
		return model.Component{}, false
	}
	c, ok := m[key]
	return c, ok
}
