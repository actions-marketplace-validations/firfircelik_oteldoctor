package graph

import (
	"sort"
	"testing"

	"github.com/firfircelik/oteldoctor/internal/model"
)

func makeConfig(receivers, processors, exporters []string) *model.CollectorConfig {
	cfg := &model.CollectorConfig{
		Receivers:  make(map[string]model.Component),
		Processors: make(map[string]model.Component),
		Exporters:  make(map[string]model.Component),
		Service: model.ServiceConfig{
			Pipelines: make(map[string]model.Pipeline),
		},
	}

	for _, id := range receivers {
		cfg.Receivers[id] = model.Component{ID: id, Kind: model.ComponentKindReceiver}
	}
	for _, id := range processors {
		cfg.Processors[id] = model.Component{ID: id, Kind: model.ComponentKindProcessor}
	}
	for _, id := range exporters {
		cfg.Exporters[id] = model.Component{ID: id, Kind: model.ComponentKindExporter}
	}

	return cfg
}

func addPipeline(cfg *model.CollectorConfig, signal string, receivers, processors, exporters []string) {
	cfg.Service.Pipelines[signal] = model.Pipeline{
		Receivers:  receivers,
		Processors: processors,
		Exporters:  exporters,
	}
}

func TestGraph_Build_OneTracesPipeline(t *testing.T) {
	cfg := makeConfig(
		[]string{"otlp"},
		[]string{"batch"},
		[]string{"debug"},
	)
	addPipeline(cfg, "traces", []string{"otlp"}, []string{"batch"}, []string{"debug"})

	g := Build(cfg)

	pl, ok := g.Pipelines["traces"]
	if !ok {
		t.Fatal("expected traces pipeline")
	}

	if pl.Signal != "traces" {
		t.Errorf("expected signal traces, got %q", pl.Signal)
	}

	if len(pl.Receivers) != 1 || pl.Receivers[0].ID != "otlp" {
		t.Errorf("expected receiver [otlp], got %v", nodeIDs(pl.Receivers))
	}
	if pl.Receivers[0].Kind != model.ComponentKindReceiver {
		t.Errorf("expected receiver kind, got %q", pl.Receivers[0].Kind)
	}
	if pl.Receivers[0].Type != "otlp" || pl.Receivers[0].Name != "" {
		t.Errorf("expected type=otlp name=\"\", got type=%q name=%q", pl.Receivers[0].Type, pl.Receivers[0].Name)
	}

	if len(pl.Processors) != 1 || pl.Processors[0].ID != "batch" {
		t.Errorf("expected processor [batch], got %v", nodeIDs(pl.Processors))
	}
	if pl.Processors[0].Kind != model.ComponentKindProcessor {
		t.Errorf("expected processor kind, got %q", pl.Processors[0].Kind)
	}

	if len(pl.Exporters) != 1 || pl.Exporters[0].ID != "debug" {
		t.Errorf("expected exporter [debug], got %v", nodeIDs(pl.Exporters))
	}
	if pl.Exporters[0].Kind != model.ComponentKindExporter {
		t.Errorf("expected exporter kind, got %q", pl.Exporters[0].Kind)
	}

	for _, n := range pl.Receivers {
		if n.Pipeline != "traces" || n.Signal != "traces" {
			t.Errorf("node %q has wrong pipeline/signal", n.ID)
		}
	}

	usage := g.UsedComponents()
	if len(usage) != 3 {
		t.Errorf("expected 3 used components, got %d", len(usage))
	}

	if _, ok := usage["otlp"]; !ok {
		t.Error("expected otlp in used components")
	}
	if _, ok := usage["batch"]; !ok {
		t.Error("expected batch in used components")
	}
	if _, ok := usage["debug"]; !ok {
		t.Error("expected debug in used components")
	}

	if len(g.UndefinedReferences()) != 0 {
		t.Errorf("expected 0 undefined references, got %d", len(g.UndefinedReferences()))
	}

	if len(g.UnusedComponents()) != 0 {
		t.Errorf("expected 0 unused components, got %v", g.UnusedComponents())
	}

	order := g.PipelineProcessorOrder("traces")
	if len(order) != 1 || order[0] != "batch" {
		t.Errorf("expected processor order [batch], got %v", order)
	}

	if !g.IsComponentDefined(model.ComponentKindReceiver, "otlp") {
		t.Error("expected otlp receiver to be defined")
	}
	if g.IsComponentDefined(model.ComponentKindExporter, "otlp") {
		t.Error("otlp should not be defined as exporter")
	}

	pipelines := g.PipelinesUsingComponent(model.ComponentKindExporter, "debug")
	if len(pipelines) != 1 || pipelines[0] != "traces" {
		t.Errorf("expected debug used in [traces], got %v", pipelines)
	}
}

func TestGraph_Build_NamedComponents(t *testing.T) {
	cfg := makeConfig(
		[]string{"otlp/datadog", "otlp/newrelic"},
		[]string{"batch/default"},
		[]string{"debug/inspect"},
	)
	addPipeline(cfg, "metrics",
		[]string{"otlp/datadog", "otlp/newrelic"},
		[]string{"batch/default"},
		[]string{"debug/inspect"},
	)

	g := Build(cfg)

	pl := g.Pipelines["metrics"]
	if len(pl.Receivers) != 2 {
		t.Fatalf("expected 2 receivers, got %d", len(pl.Receivers))
	}

	r0 := pl.Receivers[0]
	if r0.ID != "otlp/datadog" || r0.Type != "otlp" || r0.Name != "datadog" {
		t.Errorf("expected otlp/datadog, got id=%q type=%q name=%q", r0.ID, r0.Type, r0.Name)
	}

	r1 := pl.Receivers[1]
	if r1.ID != "otlp/newrelic" || r1.Type != "otlp" || r1.Name != "newrelic" {
		t.Errorf("expected otlp/newrelic, got id=%q type=%q name=%q", r1.ID, r1.Type, r1.Name)
	}

	p0 := pl.Processors[0]
	if p0.ID != "batch/default" || p0.Type != "batch" || p0.Name != "default" {
		t.Errorf("expected batch/default, got id=%q type=%q name=%q", p0.ID, p0.Type, p0.Name)
	}

	e0 := pl.Exporters[0]
	if e0.ID != "debug/inspect" || e0.Type != "debug" || e0.Name != "inspect" {
		t.Errorf("expected debug/inspect, got id=%q type=%q name=%q", e0.ID, e0.Type, e0.Name)
	}
}

func TestGraph_Build_MultiplePipelinesSharingExporter(t *testing.T) {
	cfg := makeConfig(
		[]string{"otlp"},
		[]string{"batch"},
		[]string{"debug"},
	)
	addPipeline(cfg, "traces", []string{"otlp"}, []string{"batch"}, []string{"debug"})
	addPipeline(cfg, "metrics", []string{"otlp"}, []string{}, []string{"debug"})

	g := Build(cfg)

	if len(g.Pipelines) != 2 {
		t.Fatalf("expected 2 pipelines, got %d", len(g.Pipelines))
	}

	usage := g.UsedComponents()

	uDebug, ok := usage["debug"]
	if !ok {
		t.Fatal("expected debug in used components")
	}
	if len(uDebug.Pipelines) != 2 {
		t.Errorf("expected debug used in 2 pipelines, got %v", uDebug.Pipelines)
	}
	sort.Strings(uDebug.Pipelines)
	if uDebug.Pipelines[0] != "metrics" || uDebug.Pipelines[1] != "traces" {
		t.Errorf("expected debug in [metrics traces], got %v", uDebug.Pipelines)
	}
	if len(uDebug.AsKind) != 1 || uDebug.AsKind[0] != model.ComponentKindExporter {
		t.Errorf("expected debug as [exporter], got %v", uDebug.AsKind)
	}

	pipelines := g.PipelinesUsingComponent(model.ComponentKindExporter, "debug")
	sort.Strings(pipelines)
	if len(pipelines) != 2 || pipelines[0] != "metrics" || pipelines[1] != "traces" {
		t.Errorf("expected debug used in metrics and traces, got %v", pipelines)
	}
}

func TestGraph_Build_UnusedProcessor(t *testing.T) {
	cfg := makeConfig(
		[]string{"otlp"},
		[]string{"batch", "memory_limiter"}, // memory_limiter defined but not used
		[]string{"debug"},
	)
	addPipeline(cfg, "traces", []string{"otlp"}, []string{"batch"}, []string{"debug"})

	g := Build(cfg)

	unused := g.UnusedComponents()
	if len(unused) != 1 || unused[0] != "memory_limiter" {
		t.Errorf("expected unused [memory_limiter], got %v", unused)
	}
}

func TestGraph_Build_UndefinedReferenceAsFact(t *testing.T) {
	cfg := makeConfig(
		[]string{"otlp"},
		[]string{"batch"},
		[]string{"debug"},
	)
	// Pipeline references missing_exporter which is not defined
	addPipeline(cfg, "traces", []string{"otlp"}, []string{"batch"}, []string{"debug", "missing_exporter"})

	g := Build(cfg)

	// Should not panic — graph is built regardless
	if len(g.Pipelines) != 1 {
		t.Fatalf("expected 1 pipeline, got %d", len(g.Pipelines))
	}

	undefined := g.UndefinedReferences()
	if len(undefined) != 1 {
		t.Fatalf("expected 1 undefined reference, got %d", len(undefined))
	}

	ref := undefined[0]
	if ref.ID != "missing_exporter" {
		t.Errorf("expected missing_exporter, got %q", ref.ID)
	}
	if ref.Kind != model.ComponentKindExporter {
		t.Errorf("expected exporter kind, got %q", ref.Kind)
	}
	if ref.Pipeline != "traces" {
		t.Errorf("expected traces pipeline, got %q", ref.Pipeline)
	}

	if !g.IsComponentDefined(model.ComponentKindExporter, "debug") {
		t.Error("expected debug exporter to be defined")
	}
	if g.IsComponentDefined(model.ComponentKindExporter, "missing_exporter") {
		t.Error("missing_exporter should not be defined")
	}
}

func TestGraph_Build_UndefinedReceiver(t *testing.T) {
	cfg := makeConfig(
		[]string{}, // no receivers defined
		[]string{"batch"},
		[]string{"debug"},
	)
	addPipeline(cfg, "traces", []string{"missing_rcv"}, []string{"batch"}, []string{"debug"})

	g := Build(cfg)

	undefined := g.UndefinedReferences()
	if len(undefined) != 1 {
		t.Fatalf("expected 1 undefined reference, got %d", len(undefined))
	}
	if undefined[0].ID != "missing_rcv" || undefined[0].Kind != model.ComponentKindReceiver {
		t.Errorf("expected missing_rcv receiver, got id=%q kind=%q", undefined[0].ID, undefined[0].Kind)
	}
}

func TestGraph_Build_PipelineProcessorOrder(t *testing.T) {
	cfg := makeConfig(
		[]string{"otlp"},
		[]string{"batch", "filter", "memory_limiter", "resource"},
		[]string{"debug"},
	)
	addPipeline(cfg, "traces",
		[]string{"otlp"},
		[]string{"memory_limiter", "batch", "filter", "resource"},
		[]string{"debug"},
	)

	g := Build(cfg)

	order := g.PipelineProcessorOrder("traces")
	expected := []string{"memory_limiter", "batch", "filter", "resource"}

	if len(order) != len(expected) {
		t.Fatalf("expected %d processors, got %d", len(expected), len(order))
	}

	for i, id := range expected {
		if order[i] != id {
			t.Errorf("position %d: expected %q, got %q", i, id, order[i])
		}
	}

	empty := g.PipelineProcessorOrder("nonexistent")
	if len(empty) != 0 {
		t.Errorf("expected nil for nonexistent pipeline, got %v", empty)
	}
}

func TestGraph_Build_ConnectorInReceiverExporter(t *testing.T) {
	cfg := makeConfig(
		[]string{"otlp"},
		[]string{"batch"},
		[]string{"debug"},
	)
	cfg.Connectors = map[string]model.Component{
		"count_connector": {ID: "count_connector", Kind: model.ComponentKindConnector},
	}
	addPipeline(cfg, "metrics",
		[]string{"otlp", "count_connector"},
		[]string{"batch"},
		[]string{"count_connector", "debug"},
	)

	g := Build(cfg)

	undefined := g.UndefinedReferences()
	if len(undefined) != 0 {
		t.Errorf("expected 0 undefined references, got %d: %v", len(undefined), nodeIDs(undefined))
	}

	if !g.IsComponentDefined(model.ComponentKindReceiver, "count_connector") {
		t.Error("connector used as receiver should be considered defined")
	}
	if !g.IsComponentDefined(model.ComponentKindExporter, "count_connector") {
		t.Error("connector used as exporter should be considered defined")
	}

	usage := g.UsedComponents()
	u, ok := usage["count_connector"]
	if !ok {
		t.Fatal("expected count_connector in used components")
	}
	hasReceiver := false
	hasExporter := false
	for _, k := range u.AsKind {
		if k == model.ComponentKindReceiver {
			hasReceiver = true
		}
		if k == model.ComponentKindExporter {
			hasExporter = true
		}
	}
	if !hasReceiver {
		t.Error("count_connector should be used as receiver")
	}
	if !hasExporter {
		t.Error("count_connector should be used as exporter")
	}
}

func TestGraph_Build_ServiceExtensionUsed(t *testing.T) {
	cfg := makeConfig(
		[]string{"otlp"},
		[]string{"batch"},
		[]string{"debug"},
	)
	cfg.Extensions = map[string]model.Component{
		"health_check": {ID: "health_check", Kind: model.ComponentKindExtension},
		"pprof":        {ID: "pprof", Kind: model.ComponentKindExtension},
	}
	cfg.Service.Extensions = []string{"health_check"}
	addPipeline(cfg, "traces", []string{"otlp"}, []string{"batch"}, []string{"debug"})

	g := Build(cfg)

	unused := g.UnusedComponents()
	hasHealthCheck := false
	hasPprof := false
	for _, id := range unused {
		if id == "health_check" {
			hasHealthCheck = true
		}
		if id == "pprof" {
			hasPprof = true
		}
	}
	if hasHealthCheck {
		t.Error("health_check is in service.extensions, should NOT be in unused")
	}
	if !hasPprof {
		t.Error("pprof is defined but not in service.extensions, should be in unused")
	}

	if !g.IsComponentDefined(model.ComponentKindExtension, "health_check") {
		t.Error("health_check should be defined as extension")
	}
	if !g.IsComponentDefined(model.ComponentKindExtension, "pprof") {
		t.Error("pprof should be defined as extension")
	}
}

func TestGraph_Build_IsComponentDefined(t *testing.T) {
	cfg := makeConfig(
		[]string{"otlp"},
		[]string{"batch"},
		[]string{"debug"},
	)
	cfg.Connectors = map[string]model.Component{
		"count_connector": {ID: "count_connector", Kind: model.ComponentKindConnector},
	}
	addPipeline(cfg, "traces", []string{"otlp"}, []string{"batch"}, []string{"debug"})

	g := Build(cfg)

	tests := []struct {
		kind    model.ComponentKind
		id      string
		defined bool
	}{
		{model.ComponentKindReceiver, "otlp", true},
		{model.ComponentKindProcessor, "batch", true},
		{model.ComponentKindExporter, "debug", true},
		{model.ComponentKindConnector, "count_connector", true},
		{model.ComponentKindReceiver, "count_connector", true},
		{model.ComponentKindExporter, "count_connector", true},
		{model.ComponentKindExporter, "otlp", false},
		{model.ComponentKindReceiver, "debug", false},
		{model.ComponentKindProcessor, "nonesuch", false},
	}

	for _, tt := range tests {
		got := g.IsComponentDefined(tt.kind, tt.id)
		if got != tt.defined {
			t.Errorf("IsComponentDefined(%q, %q) = %v, want %v", tt.kind, tt.id, got, tt.defined)
		}
	}
}

func TestGraph_Build_EmptyPipelines(t *testing.T) {
	cfg := makeConfig(
		[]string{"otlp"},
		[]string{"batch"},
		[]string{"debug"},
	)

	g := Build(cfg)

	if len(g.Pipelines) != 0 {
		t.Errorf("expected 0 pipelines, got %d", len(g.Pipelines))
	}
	if len(g.allNodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(g.allNodes))
	}
	if len(g.UsedComponents()) != 0 {
		t.Errorf("expected 0 used components")
	}
	if len(g.UnusedComponents()) != 3 {
		t.Errorf("expected 3 unused components, got %v", g.UnusedComponents())
	}
}

func nodeIDs(nodes []Node) []string {
	ids := make([]string, len(nodes))
	for i, n := range nodes {
		ids[i] = n.ID
	}
	return ids
}
