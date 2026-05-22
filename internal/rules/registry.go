package rules

import (
	"sort"

	"github.com/firfircelik/oteldoctor/internal/model"
)

type Registry struct {
	rules []Rule
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) Register(rule Rule) {
	r.rules = append(r.rules, rule)
}

func (r *Registry) Rules() []Rule {
	result := make([]Rule, len(r.rules))
	copy(result, r.rules)
	return result
}

func (r *Registry) RunAll(ctx RuleContext, showSuppressed bool) []model.Diagnostic {
	var all []model.Diagnostic

	for _, rule := range r.rules {
		if ctx.Policy != nil && ctx.Policy.IsRuleDisabled(rule.ID()) {
			continue
		}

		diags := rule.Check(ctx)

		for i := range diags {
			if diags[i].Severity == "" {
				diags[i].Severity = rule.DefaultSeverity()
			}
			if diags[i].Category == "" {
				diags[i].Category = rule.Category()
			}
			if diags[i].RuleID == "" {
				diags[i].RuleID = rule.ID()
			}

			if ctx.Policy != nil {
				if sev, ok := ctx.Policy.RuleSeverity(rule.ID()); ok && sev != "off" {
					if override, ok := parseSeverity(sev); ok {
						diags[i].Severity = override
					}
				}
			}
		}

		if !showSuppressed {
			filtered := diags[:0]
			for _, d := range diags {
				if _, suppressed := ctx.Policy.IsSuppressed(d.RuleID, d.Location.File); !suppressed {
					filtered = append(filtered, d)
				}
			}
			diags = filtered
		}

		all = append(all, diags...)
	}

	sort.Stable(diagnosticSlice(all))
	return all
}

func parseSeverity(s string) (model.Severity, bool) {
	switch s {
	case "critical":
		return model.SeverityCritical, true
	case "high":
		return model.SeverityHigh, true
	case "medium":
		return model.SeverityMedium, true
	case "low":
		return model.SeverityLow, true
	case "info":
		return model.SeverityInfo, true
	default:
		return "", false
	}
}

type diagnosticSlice []model.Diagnostic

var severityRank = map[model.Severity]int{
	model.SeverityCritical: 0,
	model.SeverityHigh:     1,
	model.SeverityMedium:   2,
	model.SeverityLow:      3,
	model.SeverityInfo:     4,
}

func (d diagnosticSlice) Len() int      { return len(d) }
func (d diagnosticSlice) Swap(i, j int) { d[i], d[j] = d[j], d[i] }
func (d diagnosticSlice) Less(i, j int) bool {
	ri := severityRank[d[i].Severity]
	rj := severityRank[d[j].Severity]
	if ri != rj {
		return ri < rj
	}
	return d[i].RuleID < d[j].RuleID
}
