package cli

import (
	"fmt"
	"strings"

	"github.com/firfircelik/oteldoctor/internal/rules"
	"github.com/spf13/cobra"
)

func newExplainCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "explain <RULE_ID>",
		Short: "Explain a diagnostic rule",
		Long: `explain shows detailed documentation for a specific rule, including:
- What it checks
- Why it matters
- Bad and good examples
- How to fix it`,
		Args: cobra.ExactArgs(1),
		RunE: runExplain,
	}
}

func runExplain(cmd *cobra.Command, args []string) error {
	ruleID := args[0]

	doc, ok := rules.GetRuleDoc(ruleID)
	if !ok {
		return fmt.Errorf("unknown rule ID %q. Use oteldoctor analyze to see available rules.", ruleID)
	}

	var out strings.Builder

	out.WriteString(fmt.Sprintf("Rule: %s\n", ruleID))
	out.WriteString(fmt.Sprintf("Title: %s\n", doc.Title))
	out.WriteString(fmt.Sprintf("Category: %s\n", doc.Category))
	out.WriteString(fmt.Sprintf("Default Severity: %s\n", doc.DefaultSeverity))
	out.WriteString("\n")
	out.WriteString(fmt.Sprintf("Why it matters:\n  %s\n", doc.Why))
	out.WriteString("\n")
	out.WriteString(fmt.Sprintf("Bad example:\n%s\n", indent(doc.BadExample, "  ")))
	out.WriteString("\n")
	out.WriteString(fmt.Sprintf("Good example:\n%s\n", indent(doc.GoodExample, "  ")))
	out.WriteString("\n")
	out.WriteString(fmt.Sprintf("How to fix:\n  %s\n", doc.HowToFix))

	fmt.Fprint(cmd.OutOrStdout(), out.String())
	return nil
}

func indent(s, prefix string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}
