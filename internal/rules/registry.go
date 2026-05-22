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

func (r *Registry) RunAll(ctx RuleContext) []model.Diagnostic {
	var all []model.Diagnostic

	for _, rule := range r.rules {
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
		}
		all = append(all, diags...)
	}

	sort.Stable(diagnosticSlice(all))
	return all
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
