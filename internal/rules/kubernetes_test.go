package rules

import (
	"testing"

	"github.com/firfircelik/oteldoctor/internal/graph"
	"github.com/firfircelik/oteldoctor/internal/k8s"
	"github.com/firfircelik/oteldoctor/internal/model"
)

func makeK8sContext(cfg *model.CollectorConfig, w *k8s.Workload, svc *k8s.ServiceInfo) RuleContext {
	g := graph.Build(cfg)
	return RuleContext{Config: cfg, Graph: g, Workload: w, K8sService: svc}
}

func TestMemLimiterTooClose_AtLimit(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelProcessor(cfg, "memory_limiter/default", 6)
	cfg.Processors["memory_limiter/default"] = model.Component{
		ID: "memory_limiter/default", Kind: model.ComponentKindProcessor,
		Config:   map[string]any{"limit_mib": 512, "spike_limit_mib": 128},
		Location: model.SourceLocation{File: "test.yaml", Line: 6},
	}
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, []string{"memory_limiter/default"}, []string{"debug"}, 14)

	w := &k8s.Workload{
		Kind: "Deployment",
		Name: "collector",
		Containers: []k8s.Container{
			{Limits: k8s.ResourceValue{Raw: "512Mi", MiB: 512}},
		},
	}

	ctx := makeK8sContext(cfg, w, nil)
	rule := NewMemLimiterTooCloseRule()
	diags := rule.Check(ctx)

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
}

func TestMemLimiterTooClose_Within20Percent(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	cfg.Processors["memory_limiter"] = model.Component{
		ID: "memory_limiter", Kind: model.ComponentKindProcessor,
		Config:   map[string]any{"limit_mib": 450},
		Location: model.SourceLocation{File: "test.yaml", Line: 6},
	}
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, []string{"memory_limiter"}, []string{"debug"}, 14)

	w := &k8s.Workload{
		Kind: "Deployment",
		Containers: []k8s.Container{
			{Limits: k8s.ResourceValue{Raw: "512Mi", MiB: 512}},
		},
	}

	ctx := makeK8sContext(cfg, w, nil)
	diags := NewMemLimiterTooCloseRule().Check(ctx)

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic (within 20%%), got %d", len(diags))
	}
}

func TestMemLimiterTooClose_Safe(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	cfg.Processors["memory_limiter"] = model.Component{
		ID: "memory_limiter", Kind: model.ComponentKindProcessor,
		Config:   map[string]any{"limit_mib": 200},
		Location: model.SourceLocation{File: "test.yaml", Line: 6},
	}
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, []string{"memory_limiter"}, []string{"debug"}, 14)

	w := &k8s.Workload{
		Containers: []k8s.Container{
			{Limits: k8s.ResourceValue{Raw: "512Mi", MiB: 512}},
		},
	}

	ctx := makeK8sContext(cfg, w, nil)
	diags := NewMemLimiterTooCloseRule().Check(ctx)

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics, got %d", len(diags))
	}
}

func TestGoMemLimitMissing_Detected(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, nil, []string{"debug"}, 14)

	w := &k8s.Workload{
		Containers: []k8s.Container{
			{Env: map[string]string{}},
		},
	}

	ctx := makeK8sContext(cfg, w, nil)
	diags := NewGoMemLimitMissingRule().Check(ctx)

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
}

func TestGoMemLimitMissing_Present(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, nil, []string{"debug"}, 14)

	w := &k8s.Workload{
		Containers: []k8s.Container{
			{Env: map[string]string{"GOMEMLIMIT": "400MiB"}},
		},
	}

	ctx := makeK8sContext(cfg, w, nil)
	diags := NewGoMemLimitMissingRule().Check(ctx)

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics, got %d", len(diags))
	}
}

func TestContainerResourcesMissing_Detected(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, nil, []string{"debug"}, 14)

	w := &k8s.Workload{
		Kind: "Deployment",
		Name: "collector",
		Containers: []k8s.Container{
			{Name: "collector"},
		},
	}

	ctx := makeK8sContext(cfg, w, nil)
	diags := NewContainerResourcesMissingRule().Check(ctx)

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
}

func TestContainerResourcesMissing_Present(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, nil, []string{"debug"}, 14)

	w := &k8s.Workload{
		Containers: []k8s.Container{
			{Limits: k8s.ResourceValue{Raw: "512Mi"}},
		},
	}

	ctx := makeK8sContext(cfg, w, nil)
	diags := NewContainerResourcesMissingRule().Check(ctx)

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics, got %d", len(diags))
	}
}

func TestK8sProbeMissing_Detected(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelExporter(cfg, "debug", nil, 10)
	setRelExtension(cfg, "health_check", 4)
	cfg.Extensions["health_check"] = model.Component{
		ID: "health_check", Kind: model.ComponentKindExtension,
		Config:   map[string]any{"endpoint": "0.0.0.0:13133"},
		Location: model.SourceLocation{File: "test.yaml", Line: 4},
	}
	cfg.Service.Extensions = []string{"health_check"}
	setRelPipeline(cfg, "traces", []string{"otlp"}, nil, []string{"debug"}, 14)

	w := &k8s.Workload{
		Containers: []k8s.Container{
			{Probes: k8s.ProbeInfo{HasReadiness: false, HasLiveness: false}},
		},
	}

	ctx := makeK8sContext(cfg, w, nil)
	diags := NewK8sProbeMissingRule().Check(ctx)

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
}

func TestK8sProbeMissing_NoWorkload(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelExporter(cfg, "debug", nil, 10)
	setRelExtension(cfg, "health_check", 4)
	cfg.Extensions["health_check"] = model.Component{
		ID: "health_check", Kind: model.ComponentKindExtension,
		Config: map[string]any{"endpoint": "0.0.0.0:13133"},
	}
	cfg.Service.Extensions = []string{"health_check"}
	setRelPipeline(cfg, "traces", []string{"otlp"}, nil, []string{"debug"}, 14)

	ctx := makeK8sContext(cfg, nil, nil)
	diags := NewK8sProbeMissingRule().Check(ctx)

	if len(diags) != 0 {
		t.Fatalf("expected 0 diagnostics when no workload, got %d", len(diags))
	}
}

func TestLoadBalancerService_Detected(t *testing.T) {
	cfg := relConfig()
	setRelReceiver(cfg, "otlp", 2)
	setRelExporter(cfg, "debug", nil, 10)
	setRelPipeline(cfg, "traces", []string{"otlp"}, nil, []string{"debug"}, 14)

	svc := &k8s.ServiceInfo{Name: "collector", Type: "LoadBalancer"}
	ctx := makeK8sContext(cfg, nil, svc)
	ctx.Profile = "production"

	diags := NewLoadBalancerServiceRule().Check(ctx)

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
}

func TestLoadBalancerService_Development(t *testing.T) {
	cfg := relConfig()
	svc := &k8s.ServiceInfo{Type: "LoadBalancer"}
	ctx := makeK8sContext(cfg, nil, svc)
	ctx.Profile = "development"

	diags := NewLoadBalancerServiceRule().Check(ctx)

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics in development, got %d", len(diags))
	}
}

func TestLoadBalancerService_ClusterIP(t *testing.T) {
	cfg := relConfig()
	svc := &k8s.ServiceInfo{Type: "ClusterIP"}
	ctx := makeK8sContext(cfg, nil, svc)
	ctx.Profile = "production"

	diags := NewLoadBalancerServiceRule().Check(ctx)

	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics for ClusterIP, got %d", len(diags))
	}
}
