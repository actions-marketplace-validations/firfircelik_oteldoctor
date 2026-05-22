package rules

import (
	"fmt"

	"github.com/firfircelik/oteldoctor/internal/model"
)

func fileFromCtx(ctx RuleContext) string {
	if ctx.Config == nil {
		return ""
	}
	for _, c := range ctx.Config.Receivers {
		return c.Location.File
	}
	for _, c := range ctx.Config.Processors {
		return c.Location.File
	}
	for _, c := range ctx.Config.Exporters {
		return c.Location.File
	}
	for _, c := range ctx.Config.Connectors {
		return c.Location.File
	}
	for _, c := range ctx.Config.Extensions {
		return c.Location.File
	}
	for _, p := range ctx.Config.Service.Pipelines {
		return p.Location.File
	}
	return ""
}

func pipelineLocation(cfg *model.CollectorConfig, pipelineID string) model.SourceLocation {
	if pl, ok := cfg.Service.Pipelines[pipelineID]; ok {
		return pl.Location
	}
	return model.SourceLocation{}
}

func componentLocation(cfg *model.CollectorConfig, kind model.ComponentKind, id string) model.SourceLocation {
	var c model.Component
	var ok bool
	switch kind {
	case model.ComponentKindReceiver:
		c, ok = cfg.Receivers[id]
	case model.ComponentKindProcessor:
		c, ok = cfg.Processors[id]
	case model.ComponentKindExporter:
		c, ok = cfg.Exporters[id]
	case model.ComponentKindConnector:
		c, ok = cfg.Connectors[id]
	case model.ComponentKindExtension:
		c, ok = cfg.Extensions[id]
	}
	if ok {
		return c.Location
	}
	return model.SourceLocation{}
}

// --- OTEL-STRUCT-001: Undefined receiver reference ---

type undefinedReceiverRule struct{}

func NewUndefinedReceiverRule() Rule { return &undefinedReceiverRule{} }

func (r *undefinedReceiverRule) ID() string                   { return "OTEL-STRUCT-001" }
func (r *undefinedReceiverRule) Title() string                { return "Undefined receiver reference" }
func (r *undefinedReceiverRule) Category() model.Category     { return model.CategoryStructural }
func (r *undefinedReceiverRule) DefaultSeverity() model.Severity { return model.SeverityCritical }

func (r *undefinedReceiverRule) Check(ctx RuleContext) []model.Diagnostic {
	refs := ctx.Graph.UndefinedReferences()
	var diags []model.Diagnostic

	for _, n := range refs {
		if n.Kind != model.ComponentKindReceiver {
			continue
		}
		loc := pipelineLocation(ctx.Config, n.Pipeline)
		diags = append(diags, model.Diagnostic{
			Message:  fmt.Sprintf("Pipeline %q references receiver %q which is not defined.", n.Pipeline, n.ID),
			Fix:      fmt.Sprintf("Define a receiver named %q or remove the reference.", n.ID),
			Location: loc,
		})
	}

	return diags
}

// --- OTEL-STRUCT-002: Undefined processor reference ---

type undefinedProcessorRule struct{}

func NewUndefinedProcessorRule() Rule { return &undefinedProcessorRule{} }

func (r *undefinedProcessorRule) ID() string                   { return "OTEL-STRUCT-002" }
func (r *undefinedProcessorRule) Title() string                { return "Undefined processor reference" }
func (r *undefinedProcessorRule) Category() model.Category     { return model.CategoryStructural }
func (r *undefinedProcessorRule) DefaultSeverity() model.Severity { return model.SeverityCritical }

func (r *undefinedProcessorRule) Check(ctx RuleContext) []model.Diagnostic {
	refs := ctx.Graph.UndefinedReferences()
	var diags []model.Diagnostic

	for _, n := range refs {
		if n.Kind != model.ComponentKindProcessor {
			continue
		}
		loc := pipelineLocation(ctx.Config, n.Pipeline)
		diags = append(diags, model.Diagnostic{
			Message:  fmt.Sprintf("Pipeline %q references processor %q which is not defined.", n.Pipeline, n.ID),
			Fix:      fmt.Sprintf("Define a processor named %q or remove the reference.", n.ID),
			Location: loc,
		})
	}

	return diags
}

// --- OTEL-STRUCT-003: Undefined exporter reference ---

type undefinedExporterRule struct{}

func NewUndefinedExporterRule() Rule { return &undefinedExporterRule{} }

func (r *undefinedExporterRule) ID() string                   { return "OTEL-STRUCT-003" }
func (r *undefinedExporterRule) Title() string                { return "Undefined exporter reference" }
func (r *undefinedExporterRule) Category() model.Category     { return model.CategoryStructural }
func (r *undefinedExporterRule) DefaultSeverity() model.Severity { return model.SeverityCritical }

func (r *undefinedExporterRule) Check(ctx RuleContext) []model.Diagnostic {
	refs := ctx.Graph.UndefinedReferences()
	var diags []model.Diagnostic

	for _, n := range refs {
		if n.Kind != model.ComponentKindExporter {
			continue
		}
		loc := pipelineLocation(ctx.Config, n.Pipeline)
		diags = append(diags, model.Diagnostic{
			Message:  fmt.Sprintf("Pipeline %q references exporter %q which is not defined.", n.Pipeline, n.ID),
			Fix:      fmt.Sprintf("Define an exporter named %q or remove the reference.", n.ID),
			Location: loc,
		})
	}

	return diags
}

// --- OTEL-STRUCT-004: Undefined extension reference ---

type undefinedExtensionRule struct{}

func NewUndefinedExtensionRule() Rule { return &undefinedExtensionRule{} }

func (r *undefinedExtensionRule) ID() string                   { return "OTEL-STRUCT-004" }
func (r *undefinedExtensionRule) Title() string                { return "Undefined extension reference" }
func (r *undefinedExtensionRule) Category() model.Category     { return model.CategoryStructural }
func (r *undefinedExtensionRule) DefaultSeverity() model.Severity { return model.SeverityHigh }

func (r *undefinedExtensionRule) Check(ctx RuleContext) []model.Diagnostic {
	refs := ctx.Graph.UndefinedReferences()
	var diags []model.Diagnostic

	for _, n := range refs {
		if n.Kind != model.ComponentKindExtension {
			continue
		}
		loc := model.SourceLocation{File: fileFromCtx(ctx)}
		diags = append(diags, model.Diagnostic{
			Message:  fmt.Sprintf("service.extensions references %q which is not defined in the extensions section.", n.ID),
			Fix:      fmt.Sprintf("Define an extension named %q or remove the reference.", n.ID),
			Location: loc,
		})
	}

	return diags
}

// --- OTEL-STRUCT-005: Unused receiver ---

type unusedReceiverRule struct{}

func NewUnusedReceiverRule() Rule { return &unusedReceiverRule{} }

func (r *unusedReceiverRule) ID() string                   { return "OTEL-STRUCT-005" }
func (r *unusedReceiverRule) Title() string                { return "Unused receiver" }
func (r *unusedReceiverRule) Category() model.Category     { return model.CategoryStructural }
func (r *unusedReceiverRule) DefaultSeverity() model.Severity { return model.SeverityLow }

func (r *unusedReceiverRule) Check(ctx RuleContext) []model.Diagnostic {
	var diags []model.Diagnostic

	for id, c := range ctx.Config.Receivers {
		pipes := ctx.Graph.PipelinesUsingComponent(model.ComponentKindReceiver, id)
		if len(pipes) == 0 {
			diags = append(diags, model.Diagnostic{
				Message:  fmt.Sprintf("Receiver %q is defined but not used in any pipeline.", id),
				Fix:      fmt.Sprintf("Add %q to a pipeline or remove the receiver definition.", id),
				Location: c.Location,
			})
		}
	}

	return diags
}

// --- OTEL-STRUCT-006: Unused processor ---

type unusedProcessorRule struct{}

func NewUnusedProcessorRule() Rule { return &unusedProcessorRule{} }

func (r *unusedProcessorRule) ID() string                   { return "OTEL-STRUCT-006" }
func (r *unusedProcessorRule) Title() string                { return "Unused processor" }
func (r *unusedProcessorRule) Category() model.Category     { return model.CategoryStructural }
func (r *unusedProcessorRule) DefaultSeverity() model.Severity { return model.SeverityLow }

func (r *unusedProcessorRule) Check(ctx RuleContext) []model.Diagnostic {
	var diags []model.Diagnostic

	for id, c := range ctx.Config.Processors {
		pipes := ctx.Graph.PipelinesUsingComponent(model.ComponentKindProcessor, id)
		if len(pipes) == 0 {
			diags = append(diags, model.Diagnostic{
				Message:  fmt.Sprintf("Processor %q is defined but not used in any pipeline.", id),
				Fix:      fmt.Sprintf("Add %q to a pipeline or remove the processor definition.", id),
				Location: c.Location,
			})
		}
	}

	return diags
}

// --- OTEL-STRUCT-007: Unused exporter ---

type unusedExporterRule struct{}

func NewUnusedExporterRule() Rule { return &unusedExporterRule{} }

func (r *unusedExporterRule) ID() string                   { return "OTEL-STRUCT-007" }
func (r *unusedExporterRule) Title() string                { return "Unused exporter" }
func (r *unusedExporterRule) Category() model.Category     { return model.CategoryStructural }
func (r *unusedExporterRule) DefaultSeverity() model.Severity { return model.SeverityLow }

func (r *unusedExporterRule) Check(ctx RuleContext) []model.Diagnostic {
	var diags []model.Diagnostic

	for id, c := range ctx.Config.Exporters {
		pipes := ctx.Graph.PipelinesUsingComponent(model.ComponentKindExporter, id)
		if len(pipes) == 0 {
			diags = append(diags, model.Diagnostic{
				Message:  fmt.Sprintf("Exporter %q is defined but not used in any pipeline.", id),
				Fix:      fmt.Sprintf("Add %q to a pipeline or remove the exporter definition.", id),
				Location: c.Location,
			})
		}
	}

	return diags
}

// --- OTEL-STRUCT-008: Empty pipeline ---

type emptyPipelineRule struct{}

func NewEmptyPipelineRule() Rule { return &emptyPipelineRule{} }

func (r *emptyPipelineRule) ID() string                   { return "OTEL-STRUCT-008" }
func (r *emptyPipelineRule) Title() string                { return "Empty pipeline" }
func (r *emptyPipelineRule) Category() model.Category     { return model.CategoryStructural }
func (r *emptyPipelineRule) DefaultSeverity() model.Severity { return model.SeverityHigh }

func (r *emptyPipelineRule) Check(ctx RuleContext) []model.Diagnostic {
	var diags []model.Diagnostic

	for _, pl := range ctx.Config.Service.Pipelines {
		if len(pl.Receivers) == 0 {
			diags = append(diags, model.Diagnostic{
				Message:  fmt.Sprintf("Pipeline %q has no receivers.", pl.SignalType),
				Fix:      "Add at least one receiver to the pipeline.",
				Location: pl.Location,
			})
		}
		if len(pl.Exporters) == 0 {
			diags = append(diags, model.Diagnostic{
				Message:  fmt.Sprintf("Pipeline %q has no exporters.", pl.SignalType),
				Fix:      "Add at least one exporter to the pipeline.",
				Location: pl.Location,
			})
		}
	}

	return diags
}
