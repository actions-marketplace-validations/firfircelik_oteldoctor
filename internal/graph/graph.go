package graph

import (
	"github.com/firfircelik/oteldoctor/internal/model"
)

type Node struct {
	Kind     model.ComponentKind
	ID       string
	Type     string
	Name     string
	Signal   string
	Pipeline string
}

type Usage struct {
	Pipelines []string
	AsKind    []model.ComponentKind
}

type PipelineInfo struct {
	ID         string
	Signal     string
	Receivers  []Node
	Processors []Node
	Exporters  []Node
}

type Graph struct {
	Config        *model.CollectorConfig
	Pipelines     map[string]PipelineInfo
	definedByKind map[string]map[model.ComponentKind]bool
	allNodes      []Node
}

func Build(cfg *model.CollectorConfig) *Graph {
	g := &Graph{
		Config:        cfg,
		Pipelines:     make(map[string]PipelineInfo),
		definedByKind: make(map[string]map[model.ComponentKind]bool),
	}

	for id := range cfg.Receivers {
		g.markDefined(id, model.ComponentKindReceiver)
	}
	for id := range cfg.Processors {
		g.markDefined(id, model.ComponentKindProcessor)
	}
	for id := range cfg.Exporters {
		g.markDefined(id, model.ComponentKindExporter)
	}
	for id := range cfg.Connectors {
		g.markDefined(id, model.ComponentKindConnector)
	}
	for id := range cfg.Extensions {
		g.markDefined(id, model.ComponentKindExtension)
	}

	for signal, pl := range cfg.Service.Pipelines {
		pi := PipelineInfo{
			ID:     signal,
			Signal: signal,
		}

		for _, id := range pl.Receivers {
			n := g.addNode(model.ComponentKindReceiver, id, signal)
			pi.Receivers = append(pi.Receivers, n)
		}
		for _, id := range pl.Processors {
			n := g.addNode(model.ComponentKindProcessor, id, signal)
			pi.Processors = append(pi.Processors, n)
		}
		for _, id := range pl.Exporters {
			n := g.addNode(model.ComponentKindExporter, id, signal)
			pi.Exporters = append(pi.Exporters, n)
		}

		g.Pipelines[signal] = pi
	}

	for _, id := range cfg.Service.Extensions {
		g.addNode(model.ComponentKindExtension, id, "")
	}

	return g
}

func (g *Graph) markDefined(id string, kind model.ComponentKind) {
	if g.definedByKind[id] == nil {
		g.definedByKind[id] = make(map[model.ComponentKind]bool)
	}
	g.definedByKind[id][kind] = true
}

func (g *Graph) addNode(kind model.ComponentKind, id, pipeline string) Node {
	cid, err := model.ParseComponentID(id)
	if err != nil {
		cid = model.ComponentID{Type: id}
	}

	n := Node{
		Kind:     kind,
		ID:       id,
		Type:     cid.Type,
		Name:     cid.Name,
		Signal:   pipeline,
		Pipeline: pipeline,
	}

	g.allNodes = append(g.allNodes, n)
	return n
}

func (g *Graph) UsedComponents() map[string]Usage {
	result := make(map[string]Usage)

	for _, n := range g.allNodes {
		u := result[n.ID]
		u.Pipelines = appendIfMissing(u.Pipelines, n.Pipeline)
		u.AsKind = appendKindIfMissing(u.AsKind, n.Kind)
		result[n.ID] = u
	}

	return result
}

func (g *Graph) UndefinedReferences() []Node {
	var result []Node

	for _, n := range g.allNodes {
		if !g.IsComponentDefined(n.Kind, n.ID) {
			result = append(result, n)
		}
	}

	return result
}

func (g *Graph) UnusedComponents() []string {
	used := make(map[string]bool)
	for _, n := range g.allNodes {
		used[n.ID] = true
	}

	var result []string
	for id := range g.definedByKind {
		if !used[id] {
			result = append(result, id)
		}
	}
	return result
}

func (g *Graph) PipelineProcessorOrder(pipelineID string) []string {
	pi, ok := g.Pipelines[pipelineID]
	if !ok {
		return nil
	}

	order := make([]string, len(pi.Processors))
	for i, p := range pi.Processors {
		order[i] = p.ID
	}
	return order
}

func (g *Graph) IsComponentDefined(kind model.ComponentKind, id string) bool {
	kinds, ok := g.definedByKind[id]
	if !ok {
		return false
	}

	if kinds[kind] {
		return true
	}

	switch kind {
	case model.ComponentKindReceiver, model.ComponentKindExporter:
		return kinds[model.ComponentKindConnector]
	}

	return false
}

func (g *Graph) PipelinesUsingComponent(kind model.ComponentKind, id string) []string {
	var result []string
	for _, n := range g.allNodes {
		if n.ID == id && n.Kind == kind {
			result = appendIfMissing(result, n.Pipeline)
		}
	}
	return result
}

func appendIfMissing(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}

func appendKindIfMissing(slice []model.ComponentKind, item model.ComponentKind) []model.ComponentKind {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}
