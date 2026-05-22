package rules

import (
	"testing"

	"github.com/firfircelik/oteldoctor/internal/graph"
	"github.com/firfircelik/oteldoctor/internal/model"
)

func makeTestConfig() *model.CollectorConfig {
	return &model.CollectorConfig{
		Receivers:  make(map[string]model.Component),
		Processors: make(map[string]model.Component),
		Exporters:  make(map[string]model.Component),
		Extensions: make(map[string]model.Component),
		Service: model.ServiceConfig{
			Pipelines: make(map[string]model.Pipeline),
		},
	}
}

func addReceiver(cfg *model.CollectorConfig, id string, line int) {
	cfg.Receivers[id] = model.Component{
		ID:   id,
		Kind: model.ComponentKindReceiver,
		Location: model.SourceLocation{
			File: "test.yaml",
			Line: line,
		},
	}
}

func addProcessor(cfg *model.CollectorConfig, id string, line int) {
	cfg.Processors[id] = model.Component{
		ID:   id,
		Kind: model.ComponentKindProcessor,
		Location: model.SourceLocation{
			File: "test.yaml",
			Line: line,
		},
	}
}

func addExporter(cfg *model.CollectorConfig, id string, line int) {
	cfg.Exporters[id] = model.Component{
		ID:   id,
		Kind: model.ComponentKindExporter,
		Location: model.SourceLocation{
			File: "test.yaml",
			Line: line,
		},
	}
}

func addExtension(cfg *model.CollectorConfig, id string, line int) {
	cfg.Extensions[id] = model.Component{
		ID:   id,
		Kind: model.ComponentKindExtension,
		Location: model.SourceLocation{
			File: "test.yaml",
			Line: line,
		},
	}
}

func setPipeline(cfg *model.CollectorConfig, signal string, receivers, processors, exporters []string, line int) {
	cfg.Service.Pipelines[signal] = model.Pipeline{
		SignalType: model.SignalType(signal),
		Receivers:  receivers,
		Processors: processors,
		Exporters:  exporters,
		Location: model.SourceLocation{
			File: "test.yaml",
			Line: line,
		},
	}
}

func setServiceExtensions(cfg *model.CollectorConfig, exts []string) {
	cfg.Service.Extensions = exts
}

// --- OTEL-STRUCT-001 ---

func TestUndefinedReceiverRule_NoIssue(t *testing.T) {
	cfg := makeTestConfig()
	addReceiver(cfg, "otlp", 2)
	addProcessor(cfg, "batch", 10)
	addExporter(cfg, "debug", 14)
	setPipeline(cfg, "traces", []string{"otlp"}, []string{"batch"}, []string{"debug"}, 18)

	g := graph.Build(cfg)
	rule := NewUndefinedReceiverRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics, got %d: %v", len(diags), diags)
	}
}

func TestUndefinedReceiverRule_Detected(t *testing.T) {
	cfg := makeTestConfig()
	addProcessor(cfg, "batch", 10)
	addExporter(cfg, "debug", 14)
	setPipeline(cfg, "traces", []string{"missing_rcv"}, []string{"batch"}, []string{"debug"}, 18)

	g := graph.Build(cfg)
	rule := NewUndefinedReceiverRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}

	d := diags[0]
	if d.Location.Line != 18 {
		t.Errorf("expected pipeline location line 18, got %d", d.Location.Line)
	}
	if d.Location.File != "test.yaml" {
		t.Errorf("expected file test.yaml, got %q", d.Location.File)
	}
	if !contains(d.Message, "missing_rcv") {
		t.Errorf("expected message to mention missing_rcv, got %q", d.Message)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// --- OTEL-STRUCT-002 ---

func TestUndefinedProcessorRule_NoIssue(t *testing.T) {
	cfg := makeTestConfig()
	addReceiver(cfg, "otlp", 2)
	addProcessor(cfg, "batch", 10)
	addExporter(cfg, "debug", 14)
	setPipeline(cfg, "traces", []string{"otlp"}, []string{"batch"}, []string{"debug"}, 18)

	g := graph.Build(cfg)
	rule := NewUndefinedProcessorRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics, got %d", len(diags))
	}
}

func TestUndefinedProcessorRule_Detected(t *testing.T) {
	cfg := makeTestConfig()
	addReceiver(cfg, "otlp", 2)
	addExporter(cfg, "debug", 14)
	setPipeline(cfg, "traces", []string{"otlp"}, []string{"missing_proc"}, []string{"debug"}, 18)

	g := graph.Build(cfg)
	rule := NewUndefinedProcessorRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}

	d := diags[0]
	if d.Location.Line != 18 {
		t.Errorf("expected pipeline location line 18, got %d", d.Location.Line)
	}
}

// --- OTEL-STRUCT-003 ---

func TestUndefinedExporterRule_NoIssue(t *testing.T) {
	cfg := makeTestConfig()
	addReceiver(cfg, "otlp", 2)
	addProcessor(cfg, "batch", 10)
	addExporter(cfg, "debug", 14)
	setPipeline(cfg, "traces", []string{"otlp"}, []string{"batch"}, []string{"debug"}, 18)

	g := graph.Build(cfg)
	rule := NewUndefinedExporterRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics, got %d", len(diags))
	}
}

func TestUndefinedExporterRule_Detected(t *testing.T) {
	cfg := makeTestConfig()
	addReceiver(cfg, "otlp", 2)
	addProcessor(cfg, "batch", 10)
	setPipeline(cfg, "traces", []string{"otlp"}, []string{"batch"}, []string{"missing_exp"}, 18)

	g := graph.Build(cfg)
	rule := NewUndefinedExporterRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}

	d := diags[0]
	if d.Location.Line != 18 {
		t.Errorf("expected pipeline location line 18, got %d", d.Location.Line)
	}
}

func TestUndefinedExporterRule_ConnectorNotFlagged(t *testing.T) {
	cfg := makeTestConfig()
	addReceiver(cfg, "otlp", 2)
	addProcessor(cfg, "batch", 10)
	cfg.Connectors = map[string]model.Component{
		"count_connector": {ID: "count_connector", Kind: model.ComponentKindConnector, Location: model.SourceLocation{File: "test.yaml", Line: 8}},
	}
	setPipeline(cfg, "metrics", []string{"otlp"}, []string{"batch"}, []string{"count_connector"}, 18)

	g := graph.Build(cfg)
	rule := NewUndefinedExporterRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 0 {
		t.Errorf("connector in exporters should not be flagged as undefined, got %d diags", len(diags))
	}
}

// --- OTEL-STRUCT-004 ---

func TestUndefinedExtensionRule_NoIssue(t *testing.T) {
	cfg := makeTestConfig()
	addReceiver(cfg, "otlp", 2)
	addExporter(cfg, "debug", 14)
	addExtension(cfg, "health_check", 8)
	setPipeline(cfg, "traces", []string{"otlp"}, nil, []string{"debug"}, 18)
	setServiceExtensions(cfg, []string{"health_check"})

	g := graph.Build(cfg)
	rule := NewUndefinedExtensionRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics, got %d", len(diags))
	}
}

func TestUndefinedExtensionRule_Detected(t *testing.T) {
	cfg := makeTestConfig()
	addReceiver(cfg, "otlp", 2)
	addExporter(cfg, "debug", 14)
	setPipeline(cfg, "traces", []string{"otlp"}, nil, []string{"debug"}, 18)
	setServiceExtensions(cfg, []string{"missing_ext"})

	g := graph.Build(cfg)
	rule := NewUndefinedExtensionRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}

	d := diags[0]
	if d.Location.File != "test.yaml" {
		t.Errorf("expected file test.yaml, got %q", d.Location.File)
	}
}

// --- OTEL-STRUCT-005 ---

func TestUnusedReceiverRule_NoIssue(t *testing.T) {
	cfg := makeTestConfig()
	addReceiver(cfg, "otlp", 2)
	addExporter(cfg, "debug", 14)
	setPipeline(cfg, "traces", []string{"otlp"}, nil, []string{"debug"}, 18)

	g := graph.Build(cfg)
	rule := NewUnusedReceiverRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics, got %d", len(diags))
	}
}

func TestUnusedReceiverRule_Detected(t *testing.T) {
	cfg := makeTestConfig()
	addReceiver(cfg, "otlp", 2)
	addReceiver(cfg, "jaeger", 6)
	addExporter(cfg, "debug", 14)
	setPipeline(cfg, "traces", []string{"otlp"}, nil, []string{"debug"}, 18)

	g := graph.Build(cfg)
	rule := NewUnusedReceiverRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}

	d := diags[0]
	if d.Location.Line != 6 {
		t.Errorf("expected component location line 6, got %d", d.Location.Line)
	}
}

// --- OTEL-STRUCT-006 ---

func TestUnusedProcessorRule_NoIssue(t *testing.T) {
	cfg := makeTestConfig()
	addReceiver(cfg, "otlp", 2)
	addProcessor(cfg, "batch", 10)
	addExporter(cfg, "debug", 14)
	setPipeline(cfg, "traces", []string{"otlp"}, []string{"batch"}, []string{"debug"}, 18)

	g := graph.Build(cfg)
	rule := NewUnusedProcessorRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics, got %d", len(diags))
	}
}

func TestUnusedProcessorRule_Detected(t *testing.T) {
	cfg := makeTestConfig()
	addReceiver(cfg, "otlp", 2)
	addProcessor(cfg, "batch", 10)
	addProcessor(cfg, "memory_limiter", 12)
	addExporter(cfg, "debug", 14)
	setPipeline(cfg, "traces", []string{"otlp"}, []string{"batch"}, []string{"debug"}, 18)

	g := graph.Build(cfg)
	rule := NewUnusedProcessorRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}

	d := diags[0]
	if d.Location.Line != 12 {
		t.Errorf("expected component location line 12, got %d", d.Location.Line)
	}
}

// --- OTEL-STRUCT-007 ---

func TestUnusedExporterRule_NoIssue(t *testing.T) {
	cfg := makeTestConfig()
	addReceiver(cfg, "otlp", 2)
	addExporter(cfg, "debug", 14)
	setPipeline(cfg, "traces", []string{"otlp"}, nil, []string{"debug"}, 18)

	g := graph.Build(cfg)
	rule := NewUnusedExporterRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics, got %d", len(diags))
	}
}

func TestUnusedExporterRule_Detected(t *testing.T) {
	cfg := makeTestConfig()
	addReceiver(cfg, "otlp", 2)
	addExporter(cfg, "debug", 14)
	addExporter(cfg, "logging", 16)
	setPipeline(cfg, "traces", []string{"otlp"}, nil, []string{"debug"}, 18)

	g := graph.Build(cfg)
	rule := NewUnusedExporterRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}

	d := diags[0]
	if d.Location.Line != 16 {
		t.Errorf("expected component location line 16, got %d", d.Location.Line)
	}
}

// --- OTEL-STRUCT-008 ---

func TestEmptyPipelineRule_NoIssue(t *testing.T) {
	cfg := makeTestConfig()
	addReceiver(cfg, "otlp", 2)
	addExporter(cfg, "debug", 14)
	setPipeline(cfg, "traces", []string{"otlp"}, nil, []string{"debug"}, 18)

	g := graph.Build(cfg)
	rule := NewEmptyPipelineRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics, got %d", len(diags))
	}
}

func TestEmptyPipelineRule_NoReceivers(t *testing.T) {
	cfg := makeTestConfig()
	addReceiver(cfg, "otlp", 2)
	addExporter(cfg, "debug", 14)
	setPipeline(cfg, "traces", nil, nil, []string{"debug"}, 18)

	g := graph.Build(cfg)
	rule := NewEmptyPipelineRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}

	if diags[0].Location.Line != 18 {
		t.Errorf("expected pipeline location line 18, got %d", diags[0].Location.Line)
	}
}

func TestEmptyPipelineRule_NoExporters(t *testing.T) {
	cfg := makeTestConfig()
	addReceiver(cfg, "otlp", 2)
	setPipeline(cfg, "traces", []string{"otlp"}, nil, nil, 18)

	g := graph.Build(cfg)
	rule := NewEmptyPipelineRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}

	if diags[0].Location.Line != 18 {
		t.Errorf("expected pipeline location line 18, got %d", diags[0].Location.Line)
	}
}

func TestEmptyPipelineRule_BothMissing(t *testing.T) {
	cfg := makeTestConfig()
	addReceiver(cfg, "otlp", 2)
	addExporter(cfg, "debug", 14)
	setPipeline(cfg, "empty", nil, nil, nil, 22)

	g := graph.Build(cfg)
	rule := NewEmptyPipelineRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 2 {
		t.Fatalf("expected 2 diagnostics (no receivers + no exporters), got %d", len(diags))
	}

	if diags[0].Location.Line != 22 {
		t.Errorf("expected line 22 for first diagnostic, got %d", diags[0].Location.Line)
	}
	if diags[1].Location.Line != 22 {
		t.Errorf("expected line 22 for second diagnostic, got %d", diags[1].Location.Line)
	}
}
