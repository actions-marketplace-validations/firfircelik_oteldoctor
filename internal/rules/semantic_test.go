package rules

import (
	"testing"

	"github.com/firfircelik/oteldoctor/internal/graph"
	"github.com/firfircelik/oteldoctor/internal/model"
)

func setResourceProcessor(cfg *model.CollectorConfig, id string, line int, attrs ...map[string]any) {
	cfg.Processors[id] = model.Component{
		ID: id, Kind: model.ComponentKindProcessor,
		Config: map[string]any{
			"attributes": toAnySlice(attrs),
		},
		Location: model.SourceLocation{File: "test.yaml", Line: line},
	}
}

func setAttributesProcessor(cfg *model.CollectorConfig, id string, line int, actions ...map[string]any) {
	cfg.Processors[id] = model.Component{
		ID: id, Kind: model.ComponentKindProcessor,
		Config: map[string]any{
			"actions": toAnySlice(actions),
		},
		Location: model.SourceLocation{File: "test.yaml", Line: line},
	}
}

func setTelemetryResource(cfg *model.CollectorConfig, attrs map[string]any) {
	cfg.Service.Telemetry = map[string]any{
		"resource": attrs,
	}
}

func toAnySlice(maps []map[string]any) []any {
	result := make([]any, len(maps))
	for i, m := range maps {
		result[i] = m
	}
	return result
}

func TestServiceNameMissing_NotConfigured(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, nil, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewServiceNameMissingRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if diags[0].Severity != model.SeverityLow {
		t.Errorf("expected low severity, got %q", diags[0].Severity)
	}
}

func TestServiceNameMissing_ResourceProcessor(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setResourceProcessor(cfg, "resource", 6, map[string]any{"key": "service.name", "value": "myapp", "action": "upsert"})
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, []string{"resource"}, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewServiceNameMissingRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics when service.name set via resource processor, got %d", len(diags))
	}
}

func TestServiceNameMissing_AttributesProcessor(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setAttributesProcessor(cfg, "attributes", 6, map[string]any{"key": "service.name", "action": "upsert"})
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, []string{"attributes"}, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewServiceNameMissingRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics when service.name set via attributes processor, got %d", len(diags))
	}
}

func TestServiceNameMissing_TelemetryResource(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelExporter(cfg, "debug", nil, 10)
	setTelemetryResource(cfg, map[string]any{"service.name": "myapp"})
	setRelPipeline(cfg, "traces", []string{"otlp"}, nil, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewServiceNameMissingRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics when service.name in telemetry resource, got %d", len(diags))
	}
}

func TestLegacyServiceName_AppName(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setResourceProcessor(cfg, "resource", 6, map[string]any{"key": "app_name", "value": "myapp", "action": "upsert"})
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, []string{"resource"}, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewLegacyServiceNameRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic for app_name, got %d", len(diags))
	}
	if diags[0].Location.Line != 6 {
		t.Errorf("expected line 6, got %d", diags[0].Location.Line)
	}
}

func TestLegacyServiceName_ApplicationName(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setAttributesProcessor(cfg, "attributes", 6, map[string]any{"key": "application_name", "action": "upsert"})
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, []string{"attributes"}, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewLegacyServiceNameRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic for application_name, got %d", len(diags))
	}
}

func TestLegacyServiceName_ServiceNameIsFine(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setResourceProcessor(cfg, "resource", 6, map[string]any{"key": "service.name", "value": "myapp", "action": "upsert"})
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, []string{"resource"}, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewLegacyServiceNameRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics for service.name, got %d", len(diags))
	}
}

func TestDeploymentEnvMissing_Production(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, nil, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewDeploymentEnvMissingRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g, Profile: "production"})

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if diags[0].Severity != model.SeverityMedium {
		t.Errorf("expected medium severity, got %q", diags[0].Severity)
	}
}

func TestDeploymentEnvMissing_Development(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, nil, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewDeploymentEnvMissingRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g, Profile: "development"})

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics in development, got %d", len(diags))
	}
}

func TestDeploymentEnvMissing_Configured(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setResourceProcessor(cfg, "resource", 6, map[string]any{"key": "deployment.environment", "value": "production", "action": "upsert"})
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, []string{"resource"}, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewDeploymentEnvMissingRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g, Profile: "production"})

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics when configured, got %d", len(diags))
	}
}

func TestServiceVersionMissing_Production(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, nil, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewServiceVersionMissingRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g, Profile: "production"})

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if diags[0].Severity != model.SeverityLow {
		t.Errorf("expected low severity, got %q", diags[0].Severity)
	}
}

func TestServiceVersionMissing_Development(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, nil, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewServiceVersionMissingRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g, Profile: "development"})

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics in development, got %d", len(diags))
	}
}

func TestServiceVersionMissing_Configured(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setAttributesProcessor(cfg, "attributes", 6, map[string]any{"key": "service.version", "action": "upsert"})
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, []string{"attributes"}, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewServiceVersionMissingRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g, Profile: "production"})

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics when configured, got %d", len(diags))
	}
}

func TestDeprecatedAttribute_Detected(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setResourceProcessor(cfg, "resource", 6,
		map[string]any{"key": "http.method", "value": "GET", "action": "upsert"},
		map[string]any{"key": "service.name", "value": "myapp", "action": "upsert"},
	)
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, []string{"resource"}, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewDeprecatedAttributeRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic for deprecated attribute, got %d", len(diags))
	}
	if diags[0].Severity != model.SeverityLow {
		t.Errorf("expected low severity, got %q", diags[0].Severity)
	}
	if diags[0].Location.Line != 6 {
		t.Errorf("expected line 6, got %d", diags[0].Location.Line)
	}
}

func TestDeprecatedAttribute_None(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setResourceProcessor(cfg, "resource", 6,
		map[string]any{"key": "http.request.method", "value": "GET", "action": "upsert"},
		map[string]any{"key": "url.full", "value": "/api", "action": "upsert"},
	)
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, []string{"resource"}, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewDeprecatedAttributeRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics for current attributes, got %d", len(diags))
	}
}

func TestDeprecatedAttribute_Multiple(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setAttributesProcessor(cfg, "attributes", 6,
		map[string]any{"key": "http.method", "action": "upsert"},
		map[string]any{"key": "http.status_code", "action": "upsert"},
	)
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, []string{"attributes"}, []string{"debug"}, 14)

	g := graph.Build(cfg)
	rule := NewDeprecatedAttributeRule()
	diags := rule.Check(RuleContext{Config: cfg, Graph: g})

	if len(diags) != 2 {
		t.Fatalf("expected 2 diagnostics for 2 deprecated attrs, got %d", len(diags))
	}
}
