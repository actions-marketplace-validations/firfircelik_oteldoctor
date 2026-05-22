package rules

import (
	"fmt"

	"github.com/firfircelik/oteldoctor/internal/model"
)

func collectAllKeys(cfg map[string]any) []string {
	var keys []string
	collectKeysRecursive(cfg, &keys)
	return keys
}

func collectKeysRecursive(cfg map[string]any, keys *[]string) {
	if k, ok := cfg["key"]; ok {
		if s, ok := k.(string); ok {
			*keys = append(*keys, s)
		}
	}

	for _, v := range cfg {
		if sub, ok := v.(map[string]any); ok {
			collectKeysRecursive(sub, keys)
		}
		if arr, ok := v.([]any); ok {
			for _, item := range arr {
				if m, ok := item.(map[string]any); ok {
					collectKeysRecursive(m, keys)
				}
			}
		}
	}
}

func allAttributeKeys(cfg *model.CollectorConfig) []string {
	var keys []string

	for id := range cfg.Processors {
		typ := processorType(id)
		if typ != "resource" && typ != "attributes" {
			continue
		}
		c := componentConfig(cfg, model.ComponentKindProcessor, id)
		if c == nil {
			continue
		}
		keys = append(keys, collectAllKeys(c)...)
	}

	return keys
}

func hasAttributeKey(cfg *model.CollectorConfig, target string) bool {
	for _, k := range allAttributeKeys(cfg) {
		if k == target {
			return true
		}
	}
	return false
}

func telemetryResource(cfg *model.CollectorConfig) map[string]any {
	if cfg.Service.Telemetry == nil {
		return nil
	}
	tm, ok := cfg.Service.Telemetry.(map[string]any)
	if !ok {
		return nil
	}
	res, ok := tm["resource"]
	if !ok {
		return nil
	}
	m, ok := res.(map[string]any)
	if !ok {
		return nil
	}
	return m
}

func telemetryResourceHas(cfg *model.CollectorConfig, key string) bool {
	res := telemetryResource(cfg)
	if res == nil {
		return false
	}
	_, ok := res[key]
	return ok
}

var deprecatedAttributes = map[string]string{
	"http.method":       "http.request.method",
	"http.status_code":  "http.response.status_code",
	"http.url":          "url.full",
	"http.scheme":       "url.scheme",
	"http.target":       "url.path",
	"http.host":         "server.address",
	"net.peer.ip":       "network.peer.address",
	"net.peer.name":     "server.address",
	"net.host.name":     "server.address",
}

// --- OTEL-SEM-401: service.name not configured ---

type serviceNameMissingRule struct{}

func NewServiceNameMissingRule() Rule { return &serviceNameMissingRule{} }

func (r *serviceNameMissingRule) ID() string               { return "OTEL-SEM-401" }
func (r *serviceNameMissingRule) Title() string             { return "service.name may not be configured" }
func (r *serviceNameMissingRule) Category() model.Category  { return model.CategorySemantic }
func (r *serviceNameMissingRule) DefaultSeverity() model.Severity { return model.SeverityLow }

func (r *serviceNameMissingRule) Check(ctx RuleContext) []model.Diagnostic {
	if hasAttributeKey(ctx.Config, "service.name") {
		return nil
	}
	if telemetryResourceHas(ctx.Config, "service.name") {
		return nil
	}

	return []model.Diagnostic{
		{
			Severity: model.SeverityLow,
			Message:  "service.name is not explicitly configured. The SDK or environment may set it, but without explicit configuration the service name may be inconsistent across deployments.",
			Fix:      "Consider setting service.name using a resource processor, or via the OTEL_SERVICE_NAME environment variable.",
			Location: model.SourceLocation{File: fileFromCtx(ctx)},
		},
	}
}

// --- OTEL-SEM-402: app_name/application_name used instead of service.name ---

type legacyServiceNameRule struct{}

func NewLegacyServiceNameRule() Rule { return &legacyServiceNameRule{} }

func (r *legacyServiceNameRule) ID() string               { return "OTEL-SEM-402" }
func (r *legacyServiceNameRule) Title() string             { return "Legacy service name attribute used" }
func (r *legacyServiceNameRule) Category() model.Category  { return model.CategorySemantic }
func (r *legacyServiceNameRule) DefaultSeverity() model.Severity { return model.SeverityMedium }

func (r *legacyServiceNameRule) Check(ctx RuleContext) []model.Diagnostic {
	var diags []model.Diagnostic
	legacyNames := map[string]bool{"app_name": true, "application_name": true}

	for id, c := range ctx.Config.Processors {
		typ := processorType(id)
		if typ != "resource" && typ != "attributes" {
			continue
		}

		cfg := componentConfig(ctx.Config, model.ComponentKindProcessor, id)
		if cfg == nil {
			continue
		}

		for _, k := range collectAllKeys(cfg) {
			if legacyNames[k] {
				diags = append(diags, model.Diagnostic{
					Severity: model.SeverityMedium,
					Message:  fmt.Sprintf("Processor %q sets attribute %q instead of the standard semantic convention \"service.name\". This may cause incorrect service identification in backends.", id, k),
					Fix:      "Consider replacing this with the standard \"service.name\" attribute.",
					Location: c.Location,
				})
			}
		}
	}

	return diags
}

// --- OTEL-SEM-403: deployment.environment missing in production ---

type deploymentEnvMissingRule struct{}

func NewDeploymentEnvMissingRule() Rule { return &deploymentEnvMissingRule{} }

func (r *deploymentEnvMissingRule) ID() string               { return "OTEL-SEM-403" }
func (r *deploymentEnvMissingRule) Title() string             { return "deployment.environment missing in production" }
func (r *deploymentEnvMissingRule) Category() model.Category  { return model.CategorySemantic }
func (r *deploymentEnvMissingRule) DefaultSeverity() model.Severity { return model.SeverityMedium }

func (r *deploymentEnvMissingRule) Check(ctx RuleContext) []model.Diagnostic {
	if ctx.Profile != "production" {
		return nil
	}

	if hasAttributeKey(ctx.Config, "deployment.environment") {
		return nil
	}
	if telemetryResourceHas(ctx.Config, "deployment.environment") {
		return nil
	}

	return []model.Diagnostic{
		{
			Severity: model.SeverityMedium,
			Message:  "deployment.environment is not configured in a production profile. Without it, telemetry data may be missing the environment dimension, making it harder to distinguish production from non-production data.",
			Fix:      "Consider setting deployment.environment to \"production\" using a resource processor or via the OTEL_RESOURCE_ATTRIBUTES environment variable.",
			Location: model.SourceLocation{File: fileFromCtx(ctx)},
		},
	}
}

// --- OTEL-SEM-404: service.version missing in production ---

type serviceVersionMissingRule struct{}

func NewServiceVersionMissingRule() Rule { return &serviceVersionMissingRule{} }

func (r *serviceVersionMissingRule) ID() string               { return "OTEL-SEM-404" }
func (r *serviceVersionMissingRule) Title() string             { return "service.version missing in production" }
func (r *serviceVersionMissingRule) Category() model.Category  { return model.CategorySemantic }
func (r *serviceVersionMissingRule) DefaultSeverity() model.Severity { return model.SeverityLow }

func (r *serviceVersionMissingRule) Check(ctx RuleContext) []model.Diagnostic {
	if ctx.Profile != "production" {
		return nil
	}

	if hasAttributeKey(ctx.Config, "service.version") {
		return nil
	}
	if telemetryResourceHas(ctx.Config, "service.version") {
		return nil
	}

	return []model.Diagnostic{
		{
			Severity: model.SeverityLow,
			Message:  "service.version is not configured in a production profile. Including the service version helps correlate telemetry with deployments and simplifies debugging.",
			Fix:      "Consider setting service.version using a resource processor or via the OTEL_RESOURCE_ATTRIBUTES environment variable.",
			Location: model.SourceLocation{File: fileFromCtx(ctx)},
		},
	}
}

// --- OTEL-SEM-405: deprecated semantic attribute detected ---

type deprecatedAttributeRule struct{}

func NewDeprecatedAttributeRule() Rule { return &deprecatedAttributeRule{} }

func (r *deprecatedAttributeRule) ID() string               { return "OTEL-SEM-405" }
func (r *deprecatedAttributeRule) Title() string             { return "Deprecated semantic attribute detected" }
func (r *deprecatedAttributeRule) Category() model.Category  { return model.CategorySemantic }
func (r *deprecatedAttributeRule) DefaultSeverity() model.Severity { return model.SeverityLow }

func (r *deprecatedAttributeRule) Check(ctx RuleContext) []model.Diagnostic {
	var diags []model.Diagnostic

	for id, c := range ctx.Config.Processors {
		typ := processorType(id)
		if typ != "resource" && typ != "attributes" {
			continue
		}

		cfg := componentConfig(ctx.Config, model.ComponentKindProcessor, id)
		if cfg == nil {
			continue
		}

		for _, k := range collectAllKeys(cfg) {
			replacement, deprecated := deprecatedAttributes[k]
			if !deprecated {
				continue
			}

			diags = append(diags, model.Diagnostic{
				Severity: model.SeverityLow,
				Message:  fmt.Sprintf("Processor %q uses the deprecated attribute %q. The current semantic convention uses %q instead.", id, k, replacement),
				Fix:      fmt.Sprintf("Consider renaming %q to %q.", k, replacement),
				Location: c.Location,
			})
		}
	}

	return diags
}
