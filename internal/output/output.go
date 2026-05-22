package output

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/firfircelik/oteldoctor/internal/model"
)

var severityRank = map[model.Severity]int{
	model.SeverityCritical: 0,
	model.SeverityHigh:     1,
	model.SeverityMedium:   2,
	model.SeverityLow:      3,
	model.SeverityInfo:     4,
}

type Formatter interface {
	Format(diags []model.Diagnostic) (string, error)
}

type TextFormatter struct{}

func (f *TextFormatter) Format(diags []model.Diagnostic) (string, error) {
	if len(diags) == 0 {
		return "No issues found.\n", nil
	}

	sorted := sortDiagnostics(diags)

	byFile := make(map[string][]model.Diagnostic)
	fileOrder := []string{}
	for _, d := range sorted {
		file := d.Location.File
		if file == "" {
			file = "<unknown>"
		}
		if _, ok := byFile[file]; !ok {
			fileOrder = append(fileOrder, file)
		}
		byFile[file] = append(byFile[file], d)
	}

	var out strings.Builder

	for i, file := range fileOrder {
		if i > 0 {
			out.WriteString("\n")
		}
		out.WriteString(file)
		out.WriteString("\n")

		for _, d := range byFile[file] {
			out.WriteString("\n")
			out.WriteString(strings.ToUpper(string(d.Severity)))
			out.WriteString(" ")
			out.WriteString(d.RuleID)

			if d.Location.Line > 0 {
				out.WriteString(fmt.Sprintf(" line %d", d.Location.Line))
			}

			out.WriteString("\n")
			out.WriteString(d.Message)

			if d.Fix != "" {
				out.WriteString("\nFix: ")
				out.WriteString(d.Fix)
			}
			out.WriteString("\n")
		}
	}

	out.WriteString(fmt.Sprintf("\n%d issue", len(diags)))
	if len(diags) == 1 {
		out.WriteString(" found.\n")
	} else {
		out.WriteString("s found.\n")
	}

	return out.String(), nil
}

type JSONFormatter struct{}

type jsonDiagnostic struct {
	RuleID   string          `json:"rule_id"`
	Severity string          `json:"severity"`
	Category string          `json:"category"`
	Message  string          `json:"message"`
	Fix      string          `json:"fix,omitempty"`
	Location jsonLocation    `json:"location"`
}

type jsonLocation struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

func (f *JSONFormatter) Format(diags []model.Diagnostic) (string, error) {
	sorted := sortDiagnostics(diags)

	out := make([]jsonDiagnostic, len(sorted))
	for i, d := range sorted {
		out[i] = jsonDiagnostic{
			RuleID:   d.RuleID,
			Severity: string(d.Severity),
			Category: string(d.Category),
			Message:  d.Message,
			Fix:      d.Fix,
			Location: jsonLocation{
				File:   d.Location.File,
				Line:   d.Location.Line,
				Column: d.Location.Column,
			},
		}
	}

	b, err := json.Marshal(out)
	if err != nil {
		return "", err
	}
	return string(b) + "\n", nil
}

func sortDiagnostics(diags []model.Diagnostic) []model.Diagnostic {
	sorted := make([]model.Diagnostic, len(diags))
	copy(sorted, diags)

	sort.SliceStable(sorted, func(i, j int) bool {
		a, b := sorted[i], sorted[j]

		ra := severityRank[a.Severity]
		rb := severityRank[b.Severity]
		if ra != rb {
			return ra < rb
		}

		if a.Location.File != b.Location.File {
			return a.Location.File < b.Location.File
		}

		return a.Location.Line < b.Location.Line
	})

	return sorted
}
