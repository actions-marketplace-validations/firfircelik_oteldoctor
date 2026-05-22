package rules

import (
	"testing"

	"github.com/firfircelik/oteldoctor/internal/graph"
	"github.com/firfircelik/oteldoctor/internal/model"
)

type dummyRule struct {
	ID_       string
	Title_    string
	Category_ model.Category
	Severity_ model.Severity
	Diags     []model.Diagnostic
}

func (r *dummyRule) ID() string                { return r.ID_ }
func (r *dummyRule) Title() string             { return r.Title_ }
func (r *dummyRule) Category() model.Category  { return r.Category_ }
func (r *dummyRule) DefaultSeverity() model.Severity { return r.Severity_ }
func (r *dummyRule) Check(_ RuleContext) []model.Diagnostic { return r.Diags }

func TestRegistry_Register(t *testing.T) {
	reg := NewRegistry()

	r1 := &dummyRule{ID_: "R1", Title_: "Rule One", Category_: model.CategoryReliability}
	r2 := &dummyRule{ID_: "R2", Title_: "Rule Two", Category_: model.CategorySecurity}

	reg.Register(r1)
	reg.Register(r2)

	rules := reg.Rules()
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
	if rules[0].ID() != "R1" {
		t.Errorf("expected first rule ID R1, got %q", rules[0].ID())
	}
	if rules[1].ID() != "R2" {
		t.Errorf("expected second rule ID R2, got %q", rules[1].ID())
	}
}

func TestRegistry_Rules_ReturnsCopy(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&dummyRule{ID_: "R1"})

	rules := reg.Rules()
	rules[0] = &dummyRule{ID_: "R2"}

	actual := reg.Rules()
	if actual[0].ID() != "R1" {
		t.Error("Rules() should return a copy, not the internal slice")
	}
}

func TestRegistry_RunAll_ReturnsDiagnostics(t *testing.T) {
	reg := NewRegistry()

	reg.Register(&dummyRule{
		ID_:       "OTEL-TEST-001",
		Title_:    "Dummy test rule",
		Category_: model.CategoryReliability,
		Severity_: model.SeverityMedium,
		Diags: []model.Diagnostic{
			{
				RuleID:   "OTEL-TEST-001",
				Severity: model.SeverityMedium,
				Category: model.CategoryReliability,
				Message:  "Dummy diagnostic for testing.",
			},
		},
	})

	ctx := RuleContext{
		Config: &model.CollectorConfig{},
		Graph:  &graph.Graph{},
	}

	diags := reg.RunAll(ctx)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}

	d := diags[0]
	if d.RuleID != "OTEL-TEST-001" {
		t.Errorf("expected RuleID OTEL-TEST-001, got %q", d.RuleID)
	}
	if d.Severity != model.SeverityMedium {
		t.Errorf("expected severity medium, got %q", d.Severity)
	}
	if d.Category != model.CategoryReliability {
		t.Errorf("expected category reliability, got %q", d.Category)
	}
	if d.Message != "Dummy diagnostic for testing." {
		t.Errorf("unexpected message: %q", d.Message)
	}
}

func TestRegistry_RunAll_FillsDefaults(t *testing.T) {
	reg := NewRegistry()

	reg.Register(&dummyRule{
		ID_:       "OTEL-FILL-001",
		Title_:    "Fill defaults rule",
		Category_: model.CategorySecurity,
		Severity_: model.SeverityHigh,
		Diags: []model.Diagnostic{
			{
				Message: "Diagnostic with no RuleID, Severity, or Category set.",
			},
		},
	})

	diags := reg.RunAll(RuleContext{
		Config: &model.CollectorConfig{},
		Graph:  &graph.Graph{},
	})

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}

	d := diags[0]
	if d.RuleID != "OTEL-FILL-001" {
		t.Errorf("expected RuleID to be filled as OTEL-FILL-001, got %q", d.RuleID)
	}
	if d.Severity != model.SeverityHigh {
		t.Errorf("expected severity to be filled as high, got %q", d.Severity)
	}
	if d.Category != model.CategorySecurity {
		t.Errorf("expected category to be filled as security, got %q", d.Category)
	}
}

func TestRegistry_RunAll_PreservesExplicitValues(t *testing.T) {
	reg := NewRegistry()

	reg.Register(&dummyRule{
		ID_:       "OTEL-DEFAULT",
		Title_:    "Default rule",
		Category_: model.CategoryStructural,
		Severity_: model.SeverityLow,
		Diags: []model.Diagnostic{
			{
				RuleID:   "OTEL-OVERRIDE",
				Severity: model.SeverityCritical,
				Category: model.CategorySecurity,
				Message:  "Explicit values should not be overwritten.",
			},
		},
	})

	diags := reg.RunAll(RuleContext{
		Config: &model.CollectorConfig{},
		Graph:  &graph.Graph{},
	})

	d := diags[0]
	if d.RuleID != "OTEL-OVERRIDE" {
		t.Errorf("explicit RuleID should be preserved, got %q", d.RuleID)
	}
	if d.Severity != model.SeverityCritical {
		t.Errorf("explicit severity should be preserved, got %q", d.Severity)
	}
	if d.Category != model.CategorySecurity {
		t.Errorf("explicit category should be preserved, got %q", d.Category)
	}
}

func TestRegistry_RunAll_SeveritySorting(t *testing.T) {
	reg := NewRegistry()

	reg.Register(&dummyRule{
		ID_:       "R-LOW",
		Title_:    "Low severity rule",
		Category_: model.CategoryCost,
		Severity_: model.SeverityLow,
		Diags: []model.Diagnostic{
			{Message: "low severity", Severity: model.SeverityLow},
		},
	})

	reg.Register(&dummyRule{
		ID_:       "R-HIGH",
		Title_:    "High severity rule",
		Category_: model.CategoryReliability,
		Severity_: model.SeverityHigh,
		Diags: []model.Diagnostic{
			{Message: "high severity", Severity: model.SeverityHigh},
		},
	})

	reg.Register(&dummyRule{
		ID_:       "R-CRIT",
		Title_:    "Critical severity rule",
		Category_: model.CategoryStructural,
		Severity_: model.SeverityCritical,
		Diags: []model.Diagnostic{
			{Message: "critical severity", Severity: model.SeverityCritical},
		},
	})

	diags := reg.RunAll(RuleContext{
		Config: &model.CollectorConfig{},
		Graph:  &graph.Graph{},
	})

	if len(diags) != 3 {
		t.Fatalf("expected 3 diagnostics, got %d", len(diags))
	}

	expected := []model.Severity{
		model.SeverityCritical,
		model.SeverityHigh,
		model.SeverityLow,
	}
	for i, exp := range expected {
		if diags[i].Severity != exp {
			t.Errorf("position %d: expected %q, got %q", i, exp, diags[i].Severity)
		}
	}
}

func TestRegistry_RunAll_MultipleRulesMultipleDiagnostics(t *testing.T) {
	reg := NewRegistry()

	reg.Register(&dummyRule{
		ID_:       "R-MULTI",
		Title_:    "Multi diag rule",
		Category_: model.CategorySemantic,
		Severity_: model.SeverityMedium,
		Diags: []model.Diagnostic{
			{Message: "diag 1", Severity: model.SeverityMedium},
			{Message: "diag 2", Severity: model.SeverityHigh},
		},
	})

	reg.Register(&dummyRule{
		ID_:       "R-SINGLE",
		Title_:    "Single diag rule",
		Category_: model.CategoryKubernetes,
		Severity_: model.SeverityInfo,
		Diags: []model.Diagnostic{
			{Message: "diag 3", Severity: model.SeverityInfo},
		},
	})

	diags := reg.RunAll(RuleContext{
		Config: &model.CollectorConfig{},
		Graph:  &graph.Graph{},
	})

	if len(diags) != 3 {
		t.Fatalf("expected 3 diagnostics, got %d", len(diags))
	}

	sorted := []string{string(diags[0].Severity), string(diags[1].Severity), string(diags[2].Severity)}
	expected := []string{"high", "medium", "info"}
	for i := range expected {
		if sorted[i] != expected[i] {
			t.Errorf("position %d: expected %q, got %q", i, expected[i], sorted[i])
		}
	}
}

func TestRegistry_RunAll_EmptyRegistry(t *testing.T) {
	reg := NewRegistry()

	diags := reg.RunAll(RuleContext{
		Config: &model.CollectorConfig{},
		Graph:  &graph.Graph{},
	})

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics from empty registry, got %d", len(diags))
	}
}

func TestRegistry_RunAll_RuleReceivesContext(t *testing.T) {
	cfg := &model.CollectorConfig{
		Receivers: map[string]model.Component{
			"otlp": {ID: "otlp"},
		},
	}
	g := graph.Build(cfg)

	reg := NewRegistry()

	var capturedCtx RuleContext
	reg.Register(&dummyRule{
		ID_:       "R-CTX",
		Title_:    "Context capture rule",
		Category_: model.CategoryStructural,
		Severity_: model.SeverityInfo,
		Diags:     nil, // Check is overridden below
	})

	// We need a rule that actually captures the context, so use a custom one.
	captureRule := &contextCaptureRule{}
	reg.Register(captureRule)

	diags := reg.RunAll(RuleContext{
		Config:  cfg,
		Graph:   g,
		Profile: "production",
	})

	_ = diags
	_ = capturedCtx

	if captureRule.receivedCtx.Config != cfg {
		t.Error("rule did not receive the correct Config")
	}
	if captureRule.receivedCtx.Graph != g {
		t.Error("rule did not receive the correct Graph")
	}
	if captureRule.receivedCtx.Profile != "production" {
		t.Errorf("expected profile 'production', got %q", captureRule.receivedCtx.Profile)
	}
}

type contextCaptureRule struct {
	receivedCtx RuleContext
}

func (r *contextCaptureRule) ID() string                    { return "CAPTURE" }
func (r *contextCaptureRule) Title() string                 { return "Context Capture" }
func (r *contextCaptureRule) Category() model.Category      { return model.CategoryStructural }
func (r *contextCaptureRule) DefaultSeverity() model.Severity { return model.SeverityInfo }
func (r *contextCaptureRule) Check(ctx RuleContext) []model.Diagnostic {
	r.receivedCtx = ctx
	return nil
}
