package rules

import (
	"github.com/firfircelik/oteldoctor/internal/graph"
	"github.com/firfircelik/oteldoctor/internal/model"
)

type RuleContext struct {
	Config  *model.CollectorConfig
	Graph   *graph.Graph
	Profile string
}

type Rule interface {
	ID() string
	Title() string
	Category() model.Category
	DefaultSeverity() model.Severity
	Check(ctx RuleContext) []model.Diagnostic
}
