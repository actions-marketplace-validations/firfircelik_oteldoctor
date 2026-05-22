package parser

import (
	"fmt"
	"os"

	"github.com/firfircelik/oteldoctor/internal/model"
	"gopkg.in/yaml.v3"
)

type Parser struct {
	filePath string
}

func New(filePath string) *Parser {
	return &Parser{filePath: filePath}
}

func (p *Parser) Parse() (*model.CollectorConfig, error) {
	data, err := os.ReadFile(p.filePath)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	return p.ParseBytes(data)
}

func (p *Parser) ParseBytes(data []byte) (*model.CollectorConfig, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}

	if len(doc.Content) == 0 {
		return nil, fmt.Errorf("empty configuration file")
	}

	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("configuration root must be a mapping")
	}

	config := &model.CollectorConfig{}

	for i := 0; i < len(root.Content)-1; i += 2 {
		key := root.Content[i]
		val := root.Content[i+1]

		switch key.Value {
		case "receivers":
			recvs, err := p.parseComponents(val, model.ComponentKindReceiver)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", key.Value, err)
			}
			config.Receivers = recvs

		case "processors":
			procs, err := p.parseComponents(val, model.ComponentKindProcessor)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", key.Value, err)
			}
			config.Processors = procs

		case "exporters":
			expts, err := p.parseComponents(val, model.ComponentKindExporter)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", key.Value, err)
			}
			config.Exporters = expts

		case "connectors":
			conns, err := p.parseComponents(val, model.ComponentKindConnector)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", key.Value, err)
			}
			config.Connectors = conns

		case "extensions":
			exts, err := p.parseComponents(val, model.ComponentKindExtension)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", key.Value, err)
			}
			config.Extensions = exts

		case "service":
			svc, err := p.parseService(val)
			if err != nil {
				return nil, fmt.Errorf("service: %w", err)
			}
			config.Service = svc
		}
	}

	return config, nil
}

func (p *Parser) parseComponents(node *yaml.Node, kind model.ComponentKind) (map[string]model.Component, error) {
	if node.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("expected a mapping")
	}

	components := make(map[string]model.Component)

	for i := 0; i < len(node.Content)-1; i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]

		compID, err := model.ParseComponentID(key.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid component id %q at line %d: %w", key.Value, key.Line, err)
		}

		var safeConfig any
		if err := val.Decode(&safeConfig); err != nil {
			return nil, fmt.Errorf("decoding config for %q at line %d: %w", key.Value, key.Line, err)
		}

		components[key.Value] = model.Component{
			ID:         key.Value,
			Kind:       kind,
			Name:       compID.Name,
			Config:     safeConfig,
			ConfigNode: val,
			Location: model.SourceLocation{
				File:   p.filePath,
				Line:   key.Line,
				Column: key.Column,
			},
		}
	}

	return components, nil
}

func (p *Parser) parseService(node *yaml.Node) (model.ServiceConfig, error) {
	if node.Kind != yaml.MappingNode {
		return model.ServiceConfig{}, fmt.Errorf("service section must be a mapping")
	}

	svc := model.ServiceConfig{
		Pipelines: make(map[string]model.Pipeline),
	}

	for i := 0; i < len(node.Content)-1; i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]

		switch key.Value {
		case "extensions":
			exts, err := p.parseStringArray(val)
			if err != nil {
				return model.ServiceConfig{}, fmt.Errorf("extensions: %w", err)
			}
			svc.Extensions = exts

		case "pipelines":
			pipelines, err := p.parsePipelines(val)
			if err != nil {
				return model.ServiceConfig{}, fmt.Errorf("pipelines: %w", err)
			}
			svc.Pipelines = pipelines

		case "telemetry":
			var decoded any
			if err := val.Decode(&decoded); err != nil {
				return model.ServiceConfig{}, fmt.Errorf("telemetry: %w", err)
			}
			svc.Telemetry = decoded
		}
	}

	return svc, nil
}

func (p *Parser) parsePipelines(node *yaml.Node) (map[string]model.Pipeline, error) {
	if node.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("pipelines must be a mapping")
	}

	pipelines := make(map[string]model.Pipeline)

	for i := 0; i < len(node.Content)-1; i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]

		signalName := key.Value

		pl, err := p.parsePipeline(model.SignalType(signalName), val)
		if err != nil {
			return nil, fmt.Errorf("pipeline %q at line %d: %w", signalName, key.Line, err)
		}

		pl.SignalType = model.SignalType(signalName)
		pl.Location = model.SourceLocation{
			File:   p.filePath,
			Line:   key.Line,
			Column: key.Column,
		}

		pipelines[signalName] = pl
	}

	return pipelines, nil
}

func (p *Parser) parsePipeline(signalType model.SignalType, node *yaml.Node) (model.Pipeline, error) {
	if node.Kind != yaml.MappingNode {
		return model.Pipeline{}, fmt.Errorf("pipeline definition must be a mapping")
	}

	pl := model.Pipeline{SignalType: signalType}
	pl.Receivers = []string{}
	pl.Processors = []string{}
	pl.Exporters = []string{}

	for i := 0; i < len(node.Content)-1; i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]

		switch key.Value {
		case "receivers":
			recvs, err := p.parseStringArray(val)
			if err != nil {
				return model.Pipeline{}, fmt.Errorf("receivers: %w", err)
			}
			pl.Receivers = recvs

		case "processors":
			procs, err := p.parseStringArray(val)
			if err != nil {
				return model.Pipeline{}, fmt.Errorf("processors: %w", err)
			}
			pl.Processors = procs

		case "exporters":
			expts, err := p.parseStringArray(val)
			if err != nil {
				return model.Pipeline{}, fmt.Errorf("exporters: %w", err)
			}
			pl.Exporters = expts
		}
	}

	return pl, nil
}

func (p *Parser) parseStringArray(node *yaml.Node) ([]string, error) {
	if node.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("expected a sequence, got %s", kindName(node.Kind))
	}

	var result []string
	for _, item := range node.Content {
		if item.Kind != yaml.ScalarNode {
			return nil, fmt.Errorf("expected a string value, got %s", kindName(item.Kind))
		}
		result = append(result, item.Value)
	}

	return result, nil
}

func kindName(k yaml.Kind) string {
	switch k {
	case yaml.ScalarNode:
		return "scalar"
	case yaml.SequenceNode:
		return "sequence"
	case yaml.MappingNode:
		return "mapping"
	case yaml.DocumentNode:
		return "document"
	case yaml.AliasNode:
		return "alias"
	default:
		return "unknown"
	}
}
