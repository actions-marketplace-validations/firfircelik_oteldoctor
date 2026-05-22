package rules

import (
	"fmt"

	"github.com/firfircelik/oteldoctor/internal/model"
)

func limitMiB(cfg *model.CollectorConfig) (int64, bool) {
	for id := range cfg.Processors {
		if processorType(id) != "memory_limiter" {
			continue
		}
		c := componentConfig(cfg, model.ComponentKindProcessor, id)
		if c == nil {
			continue
		}
		if v, ok := c["limit_mib"]; ok {
			switch val := v.(type) {
			case int:
				return int64(val), true
			case int64:
				return val, true
			case float64:
				return int64(val), true
			}
		}
	}
	return 0, false
}

func hasHealthCheckExtension(cfg *model.CollectorConfig) bool {
	for id := range cfg.Extensions {
		if processorType(id) == "health_check" {
			for _, eid := range cfg.Service.Extensions {
				if eid == id {
					return true
				}
			}
		}
	}
	return false
}

// --- OTEL-K8S-501: memory_limiter.limit_mib too close to container memory limit ---

type memLimiterTooCloseRule struct{}

func NewMemLimiterTooCloseRule() Rule { return &memLimiterTooCloseRule{} }

func (r *memLimiterTooCloseRule) ID() string               { return "OTEL-K8S-501" }
func (r *memLimiterTooCloseRule) Title() string             { return "memory_limiter too close to container memory limit" }
func (r *memLimiterTooCloseRule) Category() model.Category  { return model.CategoryKubernetes }
func (r *memLimiterTooCloseRule) DefaultSeverity() model.Severity { return model.SeverityMedium }

func (r *memLimiterTooCloseRule) Check(ctx RuleContext) []model.Diagnostic {
	if ctx.Workload == nil {
		return nil
	}

	for _, c := range ctx.Workload.Containers {
		if c.Limits.MiB <= 0 {
			continue
		}

		limiterMiB, ok := limitMiB(ctx.Config)
		if !ok {
			continue
		}

		if limiterMiB >= c.Limits.MiB {
			diag := model.Diagnostic{
				Severity: model.SeverityMedium,
				Message:  fmt.Sprintf("memory_limiter.limit_mib (%d) is at or above the container memory limit (%d MiB). The collector may be OOM-killed before memory_limiter activates.", limiterMiB, c.Limits.MiB),
				Fix:      fmt.Sprintf("Consider setting memory_limiter.limit_mib to at most %d MiB (80%% of the container limit) and setting GOMEMLIMIT.", c.Limits.MiB*80/100),
				Location: model.SourceLocation{File: fileFromCtx(ctx)},
			}
			return []model.Diagnostic{diag}
		}

		if limiterMiB > c.Limits.MiB*80/100 {
			diag := model.Diagnostic{
				Severity: pickSeverity(ctx.Profile, model.SeverityMedium, model.SeverityLow, model.SeverityLow),
				Message:  fmt.Sprintf("memory_limiter.limit_mib (%d) is within 20%% of the container memory limit (%d MiB). Under load, the collector may be OOM-killed before memory_limiter can react.", limiterMiB, c.Limits.MiB),
				Fix:      fmt.Sprintf("Consider reducing memory_limiter.limit_mib to at most %d MiB.", c.Limits.MiB*80/100),
				Location: model.SourceLocation{File: fileFromCtx(ctx)},
			}
			return []model.Diagnostic{diag}
		}
	}

	return nil
}

// --- OTEL-K8S-502: GOMEMLIMIT missing ---

type goMemLimitMissingRule struct{}

func NewGoMemLimitMissingRule() Rule { return &goMemLimitMissingRule{} }

func (r *goMemLimitMissingRule) ID() string               { return "OTEL-K8S-502" }
func (r *goMemLimitMissingRule) Title() string             { return "GOMEMLIMIT environment variable missing" }
func (r *goMemLimitMissingRule) Category() model.Category  { return model.CategoryKubernetes }
func (r *goMemLimitMissingRule) DefaultSeverity() model.Severity { return model.SeverityMedium }

func (r *goMemLimitMissingRule) Check(ctx RuleContext) []model.Diagnostic {
	if ctx.Workload == nil {
		return nil
	}

	for _, c := range ctx.Workload.Containers {
		if _, ok := c.Env["GOMEMLIMIT"]; ok {
			return nil
		}
	}

	return []model.Diagnostic{
		{
			Severity: model.SeverityMedium,
			Message:  "GOMEMLIMIT environment variable is not set. Without it, the Go runtime may not release memory back to the OS, leading to OOM kills under sustained load.",
			Fix:      "Consider setting GOMEMLIMIT to 80% of the container memory limit (e.g. GOMEMLIMIT=400MiB for a 512Mi container).",
			Location: model.SourceLocation{File: fileFromCtx(ctx)},
		},
	}
}

// --- OTEL-K8S-503: container resources missing ---

type containerResourcesMissingRule struct{}

func NewContainerResourcesMissingRule() Rule { return &containerResourcesMissingRule{} }

func (r *containerResourcesMissingRule) ID() string               { return "OTEL-K8S-503" }
func (r *containerResourcesMissingRule) Title() string             { return "Container resources not defined" }
func (r *containerResourcesMissingRule) Category() model.Category  { return model.CategoryKubernetes }
func (r *containerResourcesMissingRule) DefaultSeverity() model.Severity { return model.SeverityMedium }

func (r *containerResourcesMissingRule) Check(ctx RuleContext) []model.Diagnostic {
	if ctx.Workload == nil {
		return nil
	}

	for _, c := range ctx.Workload.Containers {
		if c.Limits.Raw == "" {
			return []model.Diagnostic{
				{
					Severity: model.SeverityMedium,
					Message:  fmt.Sprintf("Container %q in %s %q has no resource limits defined. Without limits, the collector may consume unbounded resources on the node.", c.Name, ctx.Workload.Kind, ctx.Workload.Name),
					Fix:      "Consider setting resources.limits.memory and resources.limits.cpu for the collector container.",
					Location: model.SourceLocation{File: fileFromCtx(ctx)},
				},
			}
		}
	}

	return nil
}

// --- OTEL-K8S-504: health_check extension exists but K8s probe missing ---

type k8sProbeMissingRule struct{}

func NewK8sProbeMissingRule() Rule { return &k8sProbeMissingRule{} }

func (r *k8sProbeMissingRule) ID() string               { return "OTEL-K8S-504" }
func (r *k8sProbeMissingRule) Title() string             { return "Kubernetes probe missing for health_check extension" }
func (r *k8sProbeMissingRule) Category() model.Category  { return model.CategoryKubernetes }
func (r *k8sProbeMissingRule) DefaultSeverity() model.Severity { return model.SeverityMedium }

func (r *k8sProbeMissingRule) Check(ctx RuleContext) []model.Diagnostic {
	if !hasHealthCheckExtension(ctx.Config) {
		return nil
	}

	if ctx.Workload == nil {
		return nil
	}

	for _, c := range ctx.Workload.Containers {
		if !c.Probes.HasReadiness || !c.Probes.HasLiveness {
			missing := []string{}
			if !c.Probes.HasReadiness {
				missing = append(missing, "readinessProbe")
			}
			if !c.Probes.HasLiveness {
				missing = append(missing, "livenessProbe")
			}
			msg := "both readinessProbe and livenessProbe"
			if len(missing) == 1 {
				msg = missing[0]
			}
			return []model.Diagnostic{
				{
					Severity: model.SeverityMedium,
					Message:  fmt.Sprintf("health_check extension is enabled but the Kubernetes workload is missing %s. The orchestrator cannot detect collector health.", msg),
					Fix:      "Add a readinessProbe and livenessProbe to the collector container, targeting the health_check endpoint.",
					Location: model.SourceLocation{File: fileFromCtx(ctx)},
				},
			}
		}
	}

	return nil
}

// --- OTEL-K8S-505: receiver service exposed as LoadBalancer in production ---

type loadBalancerServiceRule struct{}

func NewLoadBalancerServiceRule() Rule { return &loadBalancerServiceRule{} }

func (r *loadBalancerServiceRule) ID() string               { return "OTEL-K8S-505" }
func (r *loadBalancerServiceRule) Title() string             { return "Receiver exposed as LoadBalancer in production" }
func (r *loadBalancerServiceRule) Category() model.Category  { return model.CategoryKubernetes }
func (r *loadBalancerServiceRule) DefaultSeverity() model.Severity { return model.SeverityMedium }

func (r *loadBalancerServiceRule) Check(ctx RuleContext) []model.Diagnostic {
	if ctx.Profile != "production" {
		return nil
	}

	if ctx.K8sService == nil || ctx.K8sService.Type != "LoadBalancer" {
		return nil
	}

	return []model.Diagnostic{
		{
			Severity: model.SeverityMedium,
			Message:  fmt.Sprintf("Service %q is exposed as type LoadBalancer in production. This may expose the collector's receiver to the public internet.", ctx.K8sService.Name),
			Fix:      "Consider using an internal load balancer or ClusterIP with an ingress controller that provides authentication.",
			Location: model.SourceLocation{File: fileFromCtx(ctx)},
		},
	}
}
