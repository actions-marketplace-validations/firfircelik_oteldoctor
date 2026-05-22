package rules

import (
	"testing"

	"github.com/firfircelik/oteldoctor/internal/graph"
	"github.com/firfircelik/oteldoctor/internal/model"
)

func TestHighCardinalityMetric_AttributesProcessor(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelProcessor(cfg, "attributes", 6)
	cfg.Processors["attributes"] = model.Component{
		ID: "attributes", Kind: model.ComponentKindProcessor,
		Config: map[string]any{
			"actions": []any{
				map[string]any{"key": "user.id", "action": "upsert"},
				map[string]any{"key": "service.name", "action": "insert"},
			},
		},
		Location: model.SourceLocation{File: "test.yaml", Line: 6},
	}
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, []string{"attributes"}, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewHighCardinalityMetricRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic for risky attribute, got %d", len(diags))
	}
	if diags[0].Severity != model.SeverityLow {
		t.Errorf("expected low severity in development, got %q", diags[0].Severity)
	}
	if diags[0].Location.Line != 6 {
		t.Errorf("expected line 6, got %d", diags[0].Location.Line)
	}
}

func TestHighCardinalityMetric_NoRiskyAttributes(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelProcessor(cfg, "attributes", 6)
	cfg.Processors["attributes"] = model.Component{
		ID: "attributes", Kind: model.ComponentKindProcessor,
		Config: map[string]any{
			"actions": []any{
				map[string]any{"key": "service.name", "action": "upsert"},
				map[string]any{"key": "environment", "action": "insert"},
			},
		},
		Location: model.SourceLocation{File: "test.yaml", Line: 6},
	}
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, []string{"attributes"}, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewHighCardinalityMetricRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics for safe attributes, got %d", len(diags))
	}
}

func TestHighCardinalityMetric_SpanmetricsConnector(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelExporter(cfg, "debug", nil, 10)
	cfg.Connectors = map[string]model.Component{
		"spanmetrics": {
			ID: "spanmetrics", Kind: model.ComponentKindConnector,
			Config: map[string]any{
				"dimensions": []any{
					map[string]any{"name": "http.url"},
					map[string]any{"name": "service.name"},
				},
			},
			Location: model.SourceLocation{File: "test.yaml", Line: 8},
		},
	}
	setRelPipeline(cfg, "metrics", []string{"otlp"}, nil, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewHighCardinalityMetricRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic for risky spanmetrics dimension, got %d", len(diags))
	}
	if diags[0].Location.Line != 8 {
		t.Errorf("expected line 8, got %d", diags[0].Location.Line)
	}
}

func TestHighCardinalityMetric_SpanmetricsDefaultDimensions(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelExporter(cfg, "debug", nil, 10)
	cfg.Connectors = map[string]model.Component{
		"spanmetrics": {
			ID: "spanmetrics", Kind: model.ComponentKindConnector,
			Config: map[string]any{
				"dimensions": map[string]any{
					"default": []any{"http.url", "trace_id"},
				},
			},
			Location: model.SourceLocation{File: "test.yaml", Line: 8},
		},
	}
	setRelPipeline(cfg, "metrics", []string{"otlp"}, nil, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewHighCardinalityMetricRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 2 {
		t.Fatalf("expected 2 diagnostics for 2 risky dimensions, got %d", len(diags))
	}
}

func TestDynamicAttributes_Detected(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelProcessor(cfg, "transform", 6)
	cfg.Processors["transform"] = model.Component{
		ID: "transform", Kind: model.ComponentKindProcessor,
		Config: map[string]any{
			"metric_statements": []any{
				map[string]any{
					"context":    "metric",
					"statements": []any{"set(description, \"test\")", "set(unit, \"ms\")"},
				},
			},
		},
		Location: model.SourceLocation{File: "test.yaml", Line: 6},
	}
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, []string{"transform"}, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewDynamicAttributesRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic for set() statements, got %d", len(diags))
	}
	if diags[0].Location.Line != 6 {
		t.Errorf("expected line 6, got %d", diags[0].Location.Line)
	}
}

func TestDynamicAttributes_NoSetStatements(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelProcessor(cfg, "transform", 6)
	cfg.Processors["transform"] = model.Component{
		ID: "transform", Kind: model.ComponentKindProcessor,
		Config: map[string]any{
			"trace_statements": []any{
				map[string]any{
					"context":    "span",
					"statements": []any{"keep_keys(attributes, [\"service.name\"])"},
				},
			},
		},
		Location: model.SourceLocation{File: "test.yaml", Line: 6},
	}
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, []string{"transform"}, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewDynamicAttributesRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics (only keep_keys, no set()), got %d", len(diags))
	}
}

func TestSpanmetricsDimensions_NoDimensions(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelExporter(cfg, "debug", nil, 10)
	cfg.Connectors = map[string]model.Component{
		"spanmetrics": {
			ID: "spanmetrics", Kind: model.ComponentKindConnector,
			Config: map[string]any{
				"histogram": map[string]any{"explicit": map[string]any{}},
			},
			Location: model.SourceLocation{File: "test.yaml", Line: 8},
		},
	}
	setRelPipeline(cfg, "metrics", []string{"otlp", "spanmetrics"}, nil, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewSpanmetricsDimensionsRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic for missing dimensions, got %d", len(diags))
	}
}

func TestSpanmetricsDimensions_SafeDimensions(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelExporter(cfg, "debug", nil, 10)
	cfg.Connectors = map[string]model.Component{
		"spanmetrics": {
			ID: "spanmetrics", Kind: model.ComponentKindConnector,
			Config: map[string]any{
				"dimensions": []any{
					map[string]any{"name": "service.name"},
					map[string]any{"name": "http.method"},
				},
			},
			Location: model.SourceLocation{File: "test.yaml", Line: 8},
		},
	}
	setRelPipeline(cfg, "metrics", []string{"otlp", "spanmetrics"}, nil, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewSpanmetricsDimensionsRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics for safe dimensions, got %d", len(diags))
	}
}

func TestDebugExporterVolume_Production(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, nil, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewDebugExporterVolumeRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g, Profile: "production"})

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic in production, got %d", len(diags))
	}
}

func TestDebugExporterVolume_Development(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, nil, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewDebugExporterVolumeRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g, Profile: "development"})

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics in development, got %d", len(diags))
	}
}

func TestTracesNoSampling_Detected(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelProcessor(cfg, "batch", 6)
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, []string{"batch"}, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewTracesNoSamplingRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g, Profile: "production"})

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if diags[0].Severity != model.SeverityHigh {
		t.Errorf("expected high severity, got %q", diags[0].Severity)
	}
	if diags[0].Location.Line != 14 {
		t.Errorf("expected pipeline location line 14, got %d", diags[0].Location.Line)
	}
}

func TestTracesNoSampling_WithTailsampling(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelProcessor(cfg, "tailsampling", 6)
	setRelProcessor(cfg, "batch", 8)
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, []string{"tailsampling", "batch"}, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewTracesNoSamplingRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g, Profile: "production"})

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics with tailsampling, got %d", len(diags))
	}
}

func TestTracesNoSampling_WithFilter(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelProcessor(cfg, "filter", 6)
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, []string{"filter"}, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewTracesNoSamplingRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g, Profile: "production"})

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics with filter, got %d", len(diags))
	}
}

func TestTracesNoSampling_Development(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelProcessor(cfg, "batch", 6)
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, []string{"batch"}, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewTracesNoSamplingRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g, Profile: "development"})

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics in development, got %d", len(diags))
	}
}

func TestTracesNoSampling_MetricsUnaffected(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelProcessor(cfg, "batch", 6)
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "metrics", []string{"otlp"}, []string{"batch"}, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewTracesNoSamplingRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g, Profile: "production"})

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics for non-traces pipeline, got %d", len(diags))
	}
}
