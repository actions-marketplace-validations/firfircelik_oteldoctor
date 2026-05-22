package policy

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, fileName)
	os.WriteFile(p, []byte(content), 0644)
	return p
}

func TestLoad_Valid(t *testing.T) {
	p := writeTempFile(t, `
profile: staging
fail_on: medium
rules:
  OTEL-REL-101: off
  OTEL-SEC-201: high
allowed_plain_http_endpoints:
  - http://localhost:8888
allowed_high_cardinality_attributes:
  - user.id
suppressions:
  - rule: OTEL-SEC-201
    file: collector-dev.yaml
    reason: "local development"
`)

	pol, err := Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pol.Profile != "staging" {
		t.Errorf("expected profile staging, got %q", pol.Profile)
	}
	if pol.FailOn != "medium" {
		t.Errorf("expected fail_on medium, got %q", pol.FailOn)
	}
	if len(pol.Rules) != 2 {
		t.Errorf("expected 2 rule overrides, got %d", len(pol.Rules))
	}
	if len(pol.AllowedPlainHTTPEndpoints) != 1 {
		t.Errorf("expected 1 allowed endpoint, got %d", len(pol.AllowedPlainHTTPEndpoints))
	}
	if len(pol.AllowedHighCardinalityAttributes) != 1 {
		t.Errorf("expected 1 allowed attr, got %d", len(pol.AllowedHighCardinalityAttributes))
	}
	if len(pol.Suppressions) != 1 {
		t.Errorf("expected 1 suppression, got %d", len(pol.Suppressions))
	}
}

func TestLoad_Defaults(t *testing.T) {
	p := writeTempFile(t, ``)

	pol, err := Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pol.Profile != "development" {
		t.Errorf("expected default profile development, got %q", pol.Profile)
	}
	if pol.FailOn != "low" {
		t.Errorf("expected default fail_on low, got %q", pol.FailOn)
	}
}

func TestLoad_Nonexistent(t *testing.T) {
	_, err := Load("/nonexistent/path/.oteldoctor.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestDiscover_FoundInDir(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, fileName), []byte("profile: staging\n"), 0644)

	pol, err := Discover(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pol.Profile != "staging" {
		t.Errorf("expected profile staging, got %q", pol.Profile)
	}
}

func TestDiscover_FoundInParent(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, fileName), []byte("profile: production\n"), 0644)

	child := filepath.Join(dir, "sub", "deep")
	os.MkdirAll(child, 0755)

	pol, err := Discover(child)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pol.Profile != "production" {
		t.Errorf("expected profile production, got %q", pol.Profile)
	}
}

func TestDiscover_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := Discover(dir)
	if err == nil {
		t.Fatal("expected error when no policy file found")
	}
}

func TestIsRuleDisabled(t *testing.T) {
	pol := &Policy{
		Rules: map[string]string{
			"OTEL-REL-101": "off",
			"OTEL-REL-102": "high",
		},
	}

	if !pol.IsRuleDisabled("OTEL-REL-101") {
		t.Error("expected OTEL-REL-101 to be disabled")
	}
	if pol.IsRuleDisabled("OTEL-REL-102") {
		t.Error("OTEL-REL-102 should not be disabled (severity override, not off)")
	}
	if pol.IsRuleDisabled("OTEL-STRUCT-001") {
		t.Error("unreferenced rule should not be disabled")
	}
}

func TestIsRuleDisabled_NilPolicy(t *testing.T) {
	var pol *Policy
	if pol.IsRuleDisabled("anything") {
		t.Error("nil policy should not disable anything")
	}
}

func TestRuleSeverity(t *testing.T) {
	pol := &Policy{
		Rules: map[string]string{
			"OTEL-SEC-201": "low",
		},
	}

	sev, ok := pol.RuleSeverity("OTEL-SEC-201")
	if !ok {
		t.Fatal("expected severity override to exist")
	}
	if sev != "low" {
		t.Errorf("expected low, got %q", sev)
	}

	_, ok = pol.RuleSeverity("unknown")
	if ok {
		t.Error("expected no override for unknown rule")
	}
}

func TestIsEndpointAllowed(t *testing.T) {
	pol := &Policy{
		AllowedPlainHTTPEndpoints: []string{"http://localhost:8888"},
	}

	if !pol.IsEndpointAllowed("http://localhost:8888") {
		t.Error("expected endpoint to be allowed")
	}
	if pol.IsEndpointAllowed("http://remote:4318") {
		t.Error("expected endpoint to not be allowed")
	}
}

func TestIsHighCardinalityAttributeAllowed(t *testing.T) {
	pol := &Policy{
		AllowedHighCardinalityAttributes: []string{"user.id"},
	}

	if !pol.IsHighCardinalityAttributeAllowed("user.id") {
		t.Error("expected user.id to be allowed")
	}
	if pol.IsHighCardinalityAttributeAllowed("trace_id") {
		t.Error("expected trace_id to not be allowed")
	}
}

func TestIsSuppressed(t *testing.T) {
	pol := &Policy{
		Suppressions: []Suppression{
			{Rule: "OTEL-SEC-201", File: "dev.yaml", Reason: "dev"},
		},
	}

	reason, ok := pol.IsSuppressed("OTEL-SEC-201", "dev.yaml")
	if !ok {
		t.Fatal("expected suppression to match")
	}
	if reason != "dev" {
		t.Errorf("expected reason 'dev', got %q", reason)
	}

	_, ok = pol.IsSuppressed("OTEL-SEC-201", "prod.yaml")
	if ok {
		t.Error("expected no match for different file")
	}

	_, ok = pol.IsSuppressed("OTEL-XXX", "dev.yaml")
	if ok {
		t.Error("expected no match for different rule")
	}
}

func TestDefault(t *testing.T) {
	pol := Default()
	if pol.Profile != "development" {
		t.Errorf("expected development, got %q", pol.Profile)
	}
	if pol.FailOn != "low" {
		t.Errorf("expected low, got %q", pol.FailOn)
	}
}
