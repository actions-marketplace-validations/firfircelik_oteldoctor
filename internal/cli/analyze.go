package cli

import (
	"fmt"

	"github.com/firfircelik/oteldoctor/internal/graph"
	"github.com/firfircelik/oteldoctor/internal/model"
	"github.com/firfircelik/oteldoctor/internal/output"
	"github.com/firfircelik/oteldoctor/internal/parser"
	"github.com/firfircelik/oteldoctor/internal/rules"
	"github.com/spf13/cobra"
)

func newAnalyzeCmd() *cobra.Command {
	var format string
	var profile string
	var failOn string

	cmd := &cobra.Command{
		Use:   "analyze <path>",
		Short: "Analyze an OpenTelemetry Collector configuration",
		Long: `analyze parses an OpenTelemetry Collector configuration file and reports
structural, reliability, security, cost/cardinality, semantic, and
Kubernetes readiness issues.`,
		Args: cobra.ExactArgs(1),
		RunE: runAnalyze,
	}

	cmd.Flags().StringVarP(&format, "format", "f", "text", "Output format: text, json")
	cmd.Flags().StringVar(&profile, "profile", "development", "Analysis profile: development, staging, production")
	cmd.Flags().StringVar(&failOn, "fail-on", "low", "Exit with code 1 if issues are at or above: critical, high, medium, low")

	return cmd
}

var failRank = map[string]int{
	"critical": 0,
	"high":     1,
	"medium":   2,
	"low":      3,
}

func severityInt(s model.Severity) int {
	switch s {
	case model.SeverityCritical:
		return 0
	case model.SeverityHigh:
		return 1
	case model.SeverityMedium:
		return 2
	case model.SeverityLow:
		return 3
	default:
		return 4
	}
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	path := args[0]
	format, _ := cmd.Flags().GetString("format")
	profile, _ := cmd.Flags().GetString("profile")
	failOn, _ := cmd.Flags().GetString("fail-on")

	threshold, ok := failRank[failOn]
	if !ok {
		return fmt.Errorf("invalid --fail-on value %q: must be critical, high, medium, or low", failOn)
	}

	p := parser.New(path)
	cfg, err := p.Parse()
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	g := graph.Build(cfg)

	reg := rules.NewRegistry()
	for _, r := range rules.AllRules() {
		reg.Register(r)
	}

	ctx := rules.RuleContext{
		Config:  cfg,
		Graph:   g,
		Profile: profile,
	}

	diags := reg.RunAll(ctx)

	var formatter output.Formatter
	switch format {
	case "json":
		formatter = &output.JSONFormatter{}
	default:
		formatter = &output.TextFormatter{}
	}

	out, err := formatter.Format(diags)
	if err != nil {
		return fmt.Errorf("formatting output: %w", err)
	}

	fmt.Fprint(cmd.OutOrStdout(), out)

	hasIssueAtThreshold := false
	for _, d := range diags {
		if severityInt(d.Severity) <= threshold {
			hasIssueAtThreshold = true
			break
		}
	}

	if hasIssueAtThreshold {
		return &ExitError{Code: 1}
	}

	return nil
}
