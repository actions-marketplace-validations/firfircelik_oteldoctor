package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/firfircelik/oteldoctor/internal/graph"
	"github.com/firfircelik/oteldoctor/internal/parser"
	"github.com/spf13/cobra"
)

func newGraphCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "graph <collector.yaml>",
		Short: "Render the pipeline graph from a collector configuration",
		Long: `graph builds the pipeline graph from an OpenTelemetry Collector config
and renders it in the requested format.

Formats:
  mermaid  Mermaid.js flowchart (default)
  dot      Graphviz DOT format
  json     JSON nodes and edges`,
		Args: cobra.ExactArgs(1),
		RunE: runGraph,
	}

	cmd.Flags().StringVarP(&format, "format", "f", "mermaid", "Output format: mermaid, dot, json")

	return cmd
}

type jsonNode struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Kind     string `json:"kind"`
	Pipeline string `json:"pipeline"`
}

type jsonEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type jsonGraph struct {
	Pipelines []jsonPipeline `json:"pipelines"`
	Nodes     []jsonNode     `json:"nodes"`
	Edges     []jsonEdge     `json:"edges"`
}

type jsonPipeline struct {
	Signal     string   `json:"signal"`
	Receivers  []string `json:"receivers"`
	Processors []string `json:"processors"`
	Exporters  []string `json:"exporters"`
}

func runGraph(cmd *cobra.Command, args []string) error {
	path := args[0]
	format, _ := cmd.Flags().GetString("format")

	p := parser.New(path)
	cfg, err := p.Parse()
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	g := graph.Build(cfg)

	var out string
	switch format {
	case "json":
		out = renderJSONGraph(g)
	case "dot":
		out = renderDOTGraph(g)
	default:
		out = renderMermaidGraph(g)
	}

	fmt.Fprint(cmd.OutOrStdout(), out)
	return nil
}

func renderMermaidGraph(g *graph.Graph) string {
	var b strings.Builder
	b.WriteString("flowchart LR\n")

	for signal, pl := range g.Pipelines {
		b.WriteString(fmt.Sprintf("  subgraph %s [%s]\n", signal, signal))

		prevID := ""
		for _, r := range pl.Receivers {
			nid := nodeID("r", r.ID, signal)
			b.WriteString(fmt.Sprintf("    %s[%s]\n", nid, r.ID))
			if prevID != "" {
				b.WriteString(fmt.Sprintf("    %s --> %s\n", prevID, nid))
			}
			prevID = nid
		}

		for _, p := range pl.Processors {
			nid := nodeID("p", p.ID, signal)
			b.WriteString(fmt.Sprintf("    %s[%s]\n", nid, p.ID))
			if prevID != "" {
				b.WriteString(fmt.Sprintf("    %s --> %s\n", prevID, nid))
			}
			prevID = nid
		}

		for _, e := range pl.Exporters {
			nid := nodeID("e", e.ID, signal)
			b.WriteString(fmt.Sprintf("    %s[%s]\n", nid, e.ID))
			if prevID != "" {
				b.WriteString(fmt.Sprintf("    %s --> %s\n", prevID, nid))
			}
			prevID = nid
		}

		b.WriteString("  end\n")
	}

	return b.String()
}

func renderDOTGraph(g *graph.Graph) string {
	var b strings.Builder
	b.WriteString("digraph oteldoctor {\n")
	b.WriteString("  rankdir=LR;\n")
	b.WriteString("  node [shape=box];\n")

	for signal, pl := range g.Pipelines {
		b.WriteString(fmt.Sprintf("  subgraph cluster_%s {\n", signal))
		b.WriteString(fmt.Sprintf("    label=\"%s\";\n", signal))

		prevID := ""
		for _, r := range pl.Receivers {
			nid := nodeID("r", r.ID, signal)
			b.WriteString(fmt.Sprintf("    %s [label=\"%s\"];\n", nid, r.ID))
			if prevID != "" {
				b.WriteString(fmt.Sprintf("    %s -> %s;\n", prevID, nid))
			}
			prevID = nid
		}

		for _, p := range pl.Processors {
			nid := nodeID("p", p.ID, signal)
			b.WriteString(fmt.Sprintf("    %s [label=\"%s\"];\n", nid, p.ID))
			if prevID != "" {
				b.WriteString(fmt.Sprintf("    %s -> %s;\n", prevID, nid))
			}
			prevID = nid
		}

		for _, e := range pl.Exporters {
			nid := nodeID("e", e.ID, signal)
			b.WriteString(fmt.Sprintf("    %s [label=\"%s\"];\n", nid, e.ID))
			if prevID != "" {
				b.WriteString(fmt.Sprintf("    %s -> %s;\n", prevID, nid))
			}
			prevID = nid
		}

		b.WriteString("  }\n")
	}

	b.WriteString("}\n")
	return b.String()
}

func renderJSONGraph(g *graph.Graph) string {
	var jg jsonGraph

	for signal, pl := range g.Pipelines {
		jp := jsonPipeline{Signal: signal}

		prevID := ""
		for _, r := range pl.Receivers {
			nid := nodeID("r", r.ID, signal)
			jp.Receivers = append(jp.Receivers, r.ID)
			jg.Nodes = append(jg.Nodes, jsonNode{ID: nid, Type: r.Type, Kind: string(r.Kind), Pipeline: signal})
			if prevID != "" {
				jg.Edges = append(jg.Edges, jsonEdge{From: prevID, To: nid})
			}
			prevID = nid
		}

		for _, p := range pl.Processors {
			nid := nodeID("p", p.ID, signal)
			jp.Processors = append(jp.Processors, p.ID)
			jg.Nodes = append(jg.Nodes, jsonNode{ID: nid, Type: p.Type, Kind: string(p.Kind), Pipeline: signal})
			if prevID != "" {
				jg.Edges = append(jg.Edges, jsonEdge{From: prevID, To: nid})
			}
			prevID = nid
		}

		for _, e := range pl.Exporters {
			nid := nodeID("e", e.ID, signal)
			jp.Exporters = append(jp.Exporters, e.ID)
			jg.Nodes = append(jg.Nodes, jsonNode{ID: nid, Type: e.Type, Kind: string(e.Kind), Pipeline: signal})
			if prevID != "" {
				jg.Edges = append(jg.Edges, jsonEdge{From: prevID, To: nid})
			}
			prevID = nid
		}

		jg.Pipelines = append(jg.Pipelines, jp)
	}

	b, _ := json.MarshalIndent(jg, "", "  ")
	return string(b) + "\n"
}

func nodeID(prefix, id, pipeline string) string {
	safe := strings.NewReplacer("/", "_", ".", "_", "-", "_").Replace(id)
	return fmt.Sprintf("%s_%s_%s", prefix, pipeline, safe)
}
