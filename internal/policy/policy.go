package policy

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const fileName = ".oteldoctor.yaml"

type Suppression struct {
	Rule   string `yaml:"rule"`
	File   string `yaml:"file"`
	Reason string `yaml:"reason"`
}

type Policy struct {
	Profile                          string        `yaml:"profile"`
	FailOn                           string        `yaml:"fail_on"`
	Rules                            map[string]string `yaml:"rules"`
	AllowedPlainHTTPEndpoints        []string      `yaml:"allowed_plain_http_endpoints"`
	AllowedHighCardinalityAttributes []string      `yaml:"allowed_high_cardinality_attributes"`
	Suppressions                     []Suppression `yaml:"suppressions"`
}

func Default() *Policy {
	return &Policy{
		Profile: "development",
		FailOn:  "low",
	}
}

func Load(path string) (*Policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading policy file: %w", err)
	}

	var p Policy
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("invalid policy YAML: %w", err)
	}

	if p.Profile == "" {
		p.Profile = "development"
	}
	if p.FailOn == "" {
		p.FailOn = "low"
	}

	return &p, nil
}

func Discover(dir string) (*Policy, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolving path: %w", err)
	}

	for {
		p := filepath.Join(abs, fileName)
		if _, err := os.Stat(p); err == nil {
			pol, err := Load(p)
			if err == nil {
				return pol, nil
			}
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			break
		}
		abs = parent
	}

	return nil, errors.New("no .oteldoctor.yaml found in directory tree")
}

func (p *Policy) IsRuleDisabled(ruleID string) bool {
	if p == nil {
		return false
	}
	sev, ok := p.Rules[ruleID]
	return ok && sev == "off"
}

func (p *Policy) RuleSeverity(ruleID string) (string, bool) {
	if p == nil {
		return "", false
	}
	sev, ok := p.Rules[ruleID]
	return sev, ok
}

func (p *Policy) IsEndpointAllowed(ep string) bool {
	if p == nil {
		return false
	}
	for _, allowed := range p.AllowedPlainHTTPEndpoints {
		if allowed == ep {
			return true
		}
	}
	return false
}

func (p *Policy) IsHighCardinalityAttributeAllowed(attr string) bool {
	if p == nil {
		return false
	}
	for _, allowed := range p.AllowedHighCardinalityAttributes {
		if allowed == attr {
			return true
		}
	}
	return false
}

func (p *Policy) IsSuppressed(ruleID, file string) (string, bool) {
	if p == nil {
		return "", false
	}
	for _, s := range p.Suppressions {
		if s.Rule == ruleID && s.File == file {
			return s.Reason, true
		}
	}
	return "", false
}
