package model

import "gopkg.in/yaml.v3"

type ComponentKind string

const (
	ComponentKindReceiver  ComponentKind = "receiver"
	ComponentKindProcessor ComponentKind = "processor"
	ComponentKindExporter  ComponentKind = "exporter"
	ComponentKindConnector ComponentKind = "connector"
	ComponentKindExtension ComponentKind = "extension"
)

func (k ComponentKind) Valid() bool {
	switch k {
	case ComponentKindReceiver,
		ComponentKindProcessor,
		ComponentKindExporter,
		ComponentKindConnector,
		ComponentKindExtension:
		return true
	}
	return false
}

type SignalType string

const (
	SignalTypeTraces  SignalType = "traces"
	SignalTypeMetrics SignalType = "metrics"
	SignalTypeLogs    SignalType = "logs"
)

func (s SignalType) Valid() bool {
	switch s {
	case SignalTypeTraces, SignalTypeMetrics, SignalTypeLogs:
		return true
	}
	return false
}

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

func (s Severity) Valid() bool {
	switch s {
	case SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow, SeverityInfo:
		return true
	}
	return false
}

type Category string

const (
	CategoryStructural  Category = "structural"
	CategoryReliability Category = "reliability"
	CategorySecurity    Category = "security"
	CategoryCost        Category = "cost"
	CategorySemantic    Category = "semantic"
	CategoryKubernetes  Category = "kubernetes"
)

func (c Category) Valid() bool {
	switch c {
	case CategoryStructural,
		CategoryReliability,
		CategorySecurity,
		CategoryCost,
		CategorySemantic,
		CategoryKubernetes:
		return true
	}
	return false
}

type SourceLocation struct {
	File   string `yaml:"-" json:"file"`
	Line   int    `yaml:"-" json:"line"`
	Column int    `yaml:"-" json:"column"`
}

type Component struct {
	ID         string         `yaml:"-" json:"id"`
	Kind       ComponentKind  `yaml:"-" json:"kind"`
	Name       string         `yaml:"name,omitempty" json:"name,omitempty"`
	Config     any            `yaml:",inline" json:"config,omitempty"`
	ConfigNode *yaml.Node     `yaml:"-" json:"-"`
	Location   SourceLocation `yaml:"-" json:"location"`
}

type Pipeline struct {
	Receivers  []string       `yaml:"receivers" json:"receivers"`
	Processors []string       `yaml:"processors" json:"processors"`
	Exporters  []string       `yaml:"exporters" json:"exporters"`
	SignalType SignalType     `yaml:"-" json:"signal_type"`
	Location   SourceLocation `yaml:"-" json:"location"`
}

type ServiceConfig struct {
	Extensions []string            `yaml:"extensions" json:"extensions"`
	Pipelines  map[string]Pipeline `yaml:"pipelines" json:"pipelines"`
	Telemetry  any                 `yaml:"telemetry,omitempty" json:"telemetry,omitempty"`
}

type CollectorConfig struct {
	Receivers  map[string]Component `yaml:"receivers" json:"receivers"`
	Processors map[string]Component `yaml:"processors" json:"processors"`
	Exporters  map[string]Component `yaml:"exporters" json:"exporters"`
	Connectors map[string]Component `yaml:"connectors" json:"connectors"`
	Extensions map[string]Component `yaml:"extensions" json:"extensions"`
	Service    ServiceConfig        `yaml:"service" json:"service"`
}

type Diagnostic struct {
	RuleID   string         `yaml:"rule_id" json:"rule_id"`
	Severity Severity       `yaml:"severity" json:"severity"`
	Category Category       `yaml:"category" json:"category"`
	Message  string         `yaml:"message" json:"message"`
	Fix      string         `yaml:"fix,omitempty" json:"fix,omitempty"`
	Location SourceLocation `yaml:"location" json:"location"`
}
