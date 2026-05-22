package cli

import (
	"fmt"

	"github.com/firfircelik/oteldoctor/internal/autofix"
	"github.com/firfircelik/oteldoctor/internal/graph"
	"github.com/firfircelik/oteldoctor/internal/parser"
	"github.com/firfircelik/oteldoctor/internal/rules"
	"github.com/spf13/cobra"
)

func newFixCmd() *cobra.Command {
	var fixWrite bool

	cmd := &cobra.Command{
		Use:   "fix <path>",
		Short: "Auto-fix collector configuration issues",
		Long: `fix applies safe automatic corrections to an OpenTelemetry Collector configuration.

Dry-run is the default mode, showing what would change without modifying files.
Use --write to apply changes.

Currently supported fixes:
  OTEL-REL-102  Move memory_limiter to first position in processor chain`,
		Args: cobra.ExactArgs(1),
		RunE: runFix,
	}

	cmd.Flags().BoolVar(&fixWrite, "write", false, "Apply fixes to the file (default: dry-run)")
	cmd.Flags().Bool("dry-run", false, "Show what would be changed without modifying files (default)")

	return cmd
}

func runFix(cmd *cobra.Command, args []string) error {
	path := args[0]
	fixWrite, _ := cmd.Flags().GetBool("write")

	isDryRun := !fixWrite

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
		Config: cfg,
		Graph:  g,
	}

	diags := reg.RunAll(ctx, false)

	fixer := autofix.New(isDryRun)
	plans, err := fixer.Fix(path, diags)
	if err != nil {
		return fmt.Errorf("autofix error: %w", err)
	}

	if len(plans) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No auto-fixable issues found.")
		return nil
	}

	for _, plan := range plans {
		if isDryRun {
			fmt.Fprintln(cmd.OutOrStdout(), plan.Description)
			fmt.Fprint(cmd.OutOrStdout(), plan.Diff)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "Applied: %s\n", plan.Description)
		}
	}

	return nil
}
