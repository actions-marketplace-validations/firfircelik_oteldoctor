package output

import (
	"encoding/json"

	"github.com/firfircelik/oteldoctor/internal/model"
)

type sarifLog struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name           string       `json:"name"`
	InformationURI string       `json:"informationUri"`
	Rules          []sarifRule  `json:"rules"`
}

type sarifRule struct {
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	ShortDescription sarifMessage     `json:"shortDescription"`
	FullDescription  sarifMessage     `json:"fullDescription"`
	HelpURI          string           `json:"helpUri,omitempty"`
	Properties       sarifRuleProps   `json:"properties"`
}

type sarifRuleProps struct {
	Category string `json:"category"`
	Severity string `json:"severity"`
}

type sarifResult struct {
	RuleID    string         `json:"ruleId"`
	Level     string         `json:"level"`
	Message   sarifMessage   `json:"message"`
	Locations []sarifLocation `json:"locations"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
	Region           *sarifRegion          `json:"region,omitempty"`
}

type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine   int `json:"startLine,omitempty"`
	StartColumn int `json:"startColumn,omitempty"`
}

type SARIFFormatter struct{}

func (f *SARIFFormatter) Format(diags []model.Diagnostic) (string, error) {
	sorted := sortDiagnostics(diags)

	ruleSet := map[string]bool{}
	for _, d := range sorted {
		ruleSet[d.RuleID] = true
	}

	var rules []sarifRule
	for id := range ruleSet {
		sev := "warning"
		cat := "unknown"
		for _, d := range sorted {
			if d.RuleID == id {
				sev = string(d.Severity)
				cat = string(d.Category)
				break
			}
		}
		rules = append(rules, sarifRule{
			ID:   id,
			Name: id,
			ShortDescription: sarifMessage{Text: id},
			FullDescription:  sarifMessage{Text: id},
			Properties: sarifRuleProps{
				Category: cat,
				Severity: sev,
			},
		})
	}

	var results []sarifResult
	for _, d := range sorted {
		level := severityToSARIFLevel(d.Severity)

		var locs []sarifLocation
		if d.Location.File != "" || d.Location.Line > 0 {
			region := &sarifRegion{}
			hasRegion := false
			if d.Location.Line > 0 {
				region.StartLine = d.Location.Line
				hasRegion = true
			}
			if d.Location.Column > 0 {
				region.StartColumn = d.Location.Column
				hasRegion = true
			}

			var r *sarifRegion
			if hasRegion {
				r = region
			}

			locs = append(locs, sarifLocation{
				PhysicalLocation: sarifPhysicalLocation{
					ArtifactLocation: sarifArtifactLocation{
						URI: d.Location.File,
					},
					Region: r,
				},
			})
		}

		results = append(results, sarifResult{
			RuleID:  d.RuleID,
			Level:   level,
			Message: sarifMessage{Text: d.Message},
			Locations: locs,
		})
	}

	log := sarifLog{
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs: []sarifRun{
			{
				Tool: sarifTool{
					Driver: sarifDriver{
						Name:           "oteldoctor",
						InformationURI: "https://github.com/firfircelik/oteldoctor",
						Rules:          rules,
					},
				},
				Results: results,
			},
		},
	}

	b, err := json.MarshalIndent(log, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b) + "\n", nil
}

func severityToSARIFLevel(s model.Severity) string {
	switch s {
	case model.SeverityCritical, model.SeverityHigh:
		return "error"
	case model.SeverityMedium, model.SeverityLow:
		return "warning"
	default:
		return "note"
	}
}
