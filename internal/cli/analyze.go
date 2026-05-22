package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newAnalyzeCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "analyze <path>",
		Short: "Analyze an OpenTelemetry Collector configuration",
		Long: `analyze parses an OpenTelemetry Collector configuration file and reports
structural, reliability, security, cost/cardinality, semantic, and
Kubernetes readiness issues.`,
		Args: cobra.ExactArgs(1),
		RunE: runAnalyze,
	}

	cmd.Flags().StringVarP(&format, "format", "f", "text", "Output format: text, json, sarif")

	return cmd
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	path := args[0]
	_ = path

	fmt.Fprintln(cmd.OutOrStdout(), "not implemented")
	return nil
}
