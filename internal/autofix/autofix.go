package autofix

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/firfircelik/oteldoctor/internal/model"
	"gopkg.in/yaml.v3"
)

type FixPlan struct {
	FilePath    string
	Description string
	Original    []byte
	Modified    []byte
	Diff        string
	Applied     bool
}

type Fixer struct {
	dryRun bool
}

func New(dryRun bool) *Fixer {
	return &Fixer{dryRun: dryRun}
}

func (f *Fixer) Fix(filePath string, diags []model.Diagnostic) ([]FixPlan, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var plans []FixPlan

	for _, d := range diags {
		switch d.RuleID {
		case "OTEL-REL-102":
			plan, err := f.fixMemLimiterOrder(filePath, data)
			if err != nil {
				return nil, fmt.Errorf("planning fix %s: %w", d.RuleID, err)
			}
			if plan != nil {
				plans = append(plans, *plan)
			}
		}
	}

	for i := range plans {
		if f.dryRun {
			plans[i].Diff = unifiedDiff(plans[i].FilePath, plans[i].Original, plans[i].Modified)
		} else {
			if err := os.WriteFile(plans[i].FilePath, plans[i].Modified, 0644); err != nil {
				return nil, fmt.Errorf("writing file: %w", err)
			}
			plans[i].Applied = true
		}
	}

	return plans, nil
}

func (f *Fixer) fixMemLimiterOrder(filePath string, data []byte) (*FixPlan, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	if len(doc.Content) == 0 {
		return nil, nil
	}

	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, nil
	}

	serviceVal := findValue(root, "service")
	if serviceVal == nil {
		return nil, nil
	}
	pipelinesVal := findValue(serviceVal, "pipelines")
	if pipelinesVal == nil || pipelinesVal.Kind != yaml.MappingNode {
		return nil, nil
	}

	modified := false

	for i := 0; i < len(pipelinesVal.Content)-1; i += 2 {
		pipelineVal := pipelinesVal.Content[i+1]
		if pipelineVal.Kind != yaml.MappingNode {
			continue
		}

		processorsVal := findValue(pipelineVal, "processors")
		if processorsVal == nil || processorsVal.Kind != yaml.SequenceNode {
			continue
		}

		memIdx := -1
		for idx, item := range processorsVal.Content {
			if item.Kind == yaml.ScalarNode && processorType(item.Value) == "memory_limiter" {
				memIdx = idx
				break
			}
		}

		if memIdx <= 0 {
			continue
		}

		node := processorsVal.Content[memIdx]
		processorsVal.Content = append(
			processorsVal.Content[:memIdx],
			processorsVal.Content[memIdx+1:]...,
		)
		processorsVal.Content = append(
			[]*yaml.Node{node},
			processorsVal.Content...,
		)
		modified = true
	}

	if !modified {
		return nil, nil
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(&doc); err != nil {
		return nil, fmt.Errorf("encoding modified YAML: %w", err)
	}
	enc.Close()

	return &FixPlan{
		FilePath:    filePath,
		Description: "Move memory_limiter to first position in processor chain",
		Original:    data,
		Modified:    buf.Bytes(),
	}, nil
}

func findValue(node *yaml.Node, key string) *yaml.Node {
	if node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(node.Content)-1; i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

func processorType(id string) string {
	if idx := strings.Index(id, "/"); idx >= 0 {
		return id[:idx]
	}
	return id
}

func unifiedDiff(file string, orig, mod []byte) string {
	origLines := strings.SplitAfter(string(orig), "\n")
	modLines := strings.SplitAfter(string(mod), "\n")

	if strings.HasSuffix(string(orig), "\n") {
		origLines = origLines[:len(origLines)-1]
	}
	if strings.HasSuffix(string(mod), "\n") {
		modLines = modLines[:len(modLines)-1]
	}

	var out strings.Builder
	out.WriteString(fmt.Sprintf("--- %s\n", file))
	out.WriteString(fmt.Sprintf("+++ %s\n", file))
	out.WriteString(fmt.Sprintf("@@ -1,%d +1,%d @@\n", len(origLines), len(modLines)))

	i, j := 0, 0
	for i < len(origLines) || j < len(modLines) {
		if i < len(origLines) && j < len(modLines) && origLines[i] == modLines[j] {
			out.WriteString(" " + origLines[i])
			i++
			j++
		} else if j < len(modLines) {
			out.WriteString("+" + modLines[j])
			j++
			if i < len(origLines) && j < len(modLines) && origLines[i] == modLines[j] {
				continue
			}
			if i < len(origLines) {
				out.WriteString("-" + origLines[i])
				i++
			}
		} else {
			out.WriteString("-" + origLines[i])
			i++
		}
	}

	return out.String()
}
