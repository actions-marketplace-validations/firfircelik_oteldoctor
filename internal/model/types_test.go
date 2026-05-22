package model

import (
	"testing"
)

func TestComponentKind_Valid(t *testing.T) {
	tests := []struct {
		kind  ComponentKind
		valid bool
	}{
		{ComponentKindReceiver, true},
		{ComponentKindProcessor, true},
		{ComponentKindExporter, true},
		{ComponentKindConnector, true},
		{ComponentKindExtension, true},
		{ComponentKind("invalid"), false},
		{ComponentKind(""), false},
	}

	for _, tt := range tests {
		got := tt.kind.Valid()
		if got != tt.valid {
			t.Errorf("ComponentKind(%q).Valid() = %v, want %v", tt.kind, got, tt.valid)
		}
	}
}

func TestSignalType_Valid(t *testing.T) {
	tests := []struct {
		sig   SignalType
		valid bool
	}{
		{SignalTypeTraces, true},
		{SignalTypeMetrics, true},
		{SignalTypeLogs, true},
		{SignalType("events"), false},
		{SignalType(""), false},
	}

	for _, tt := range tests {
		got := tt.sig.Valid()
		if got != tt.valid {
			t.Errorf("SignalType(%q).Valid() = %v, want %v", tt.sig, got, tt.valid)
		}
	}
}

func TestSeverity_Valid(t *testing.T) {
	tests := []struct {
		sev   Severity
		valid bool
	}{
		{SeverityCritical, true},
		{SeverityHigh, true},
		{SeverityMedium, true},
		{SeverityLow, true},
		{SeverityInfo, true},
		{Severity("nuclear"), false},
		{Severity(""), false},
	}

	for _, tt := range tests {
		got := tt.sev.Valid()
		if got != tt.valid {
			t.Errorf("Severity(%q).Valid() = %v, want %v", tt.sev, got, tt.valid)
		}
	}
}

func TestCategory_Valid(t *testing.T) {
	tests := []struct {
		cat   Category
		valid bool
	}{
		{CategoryStructural, true},
		{CategoryReliability, true},
		{CategorySecurity, true},
		{CategoryCost, true},
		{CategorySemantic, true},
		{CategoryKubernetes, true},
		{Category("aesthetics"), false},
		{Category(""), false},
	}

	for _, tt := range tests {
		got := tt.cat.Valid()
		if got != tt.valid {
			t.Errorf("Category(%q).Valid() = %v, want %v", tt.cat, got, tt.valid)
		}
	}
}
