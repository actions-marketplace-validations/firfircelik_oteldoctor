package rules

import (
	"github.com/firfircelik/oteldoctor/internal/graph"
	"github.com/firfircelik/oteldoctor/internal/k8s"
	"github.com/firfircelik/oteldoctor/internal/model"
	"github.com/firfircelik/oteldoctor/internal/policy"
)

type RuleContext struct {
	Config     *model.CollectorConfig
	Graph      *graph.Graph
	Profile    string
	Policy     *policy.Policy
	Workload   *k8s.Workload
	K8sService *k8s.ServiceInfo
}

type Rule interface {
	ID() string
	Title() string
	Category() model.Category
	DefaultSeverity() model.Severity
	Check(ctx RuleContext) []model.Diagnostic
}
