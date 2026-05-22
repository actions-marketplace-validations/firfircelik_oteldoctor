package k8s

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseWorkload_Deployment(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "deploy.yaml")
	os.WriteFile(f, []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: otel-collector
spec:
  template:
    spec:
      containers:
        - name: otel-collector
          image: otel/opentelemetry-collector:0.102.0
          env:
            - name: GOMEMLIMIT
              value: "400MiB"
          resources:
            limits:
              memory: "512Mi"
              cpu: "500m"
          readinessProbe:
            httpGet:
              path: /
              port: 13133
          livenessProbe:
            httpGet:
              path: /
              port: 13133
          volumeMounts:
            - name: config
              mountPath: /etc/otel
`), 0644)

	w, err := ParseWorkload(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Kind != "Deployment" {
		t.Errorf("expected Deployment, got %q", w.Kind)
	}
	if w.Name != "otel-collector" {
		t.Errorf("expected otel-collector, got %q", w.Name)
	}
	if len(w.Containers) != 1 {
		t.Fatalf("expected 1 container, got %d", len(w.Containers))
	}

	c := w.Containers[0]
	if c.Env["GOMEMLIMIT"] != "400MiB" {
		t.Errorf("expected GOMEMLIMIT=400MiB, got %q", c.Env["GOMEMLIMIT"])
	}
	if c.Limits.MiB != 512 {
		t.Errorf("expected 512 MiB memory limit, got %d", c.Limits.MiB)
	}
	if c.Limits.CPU != "500m" {
		t.Errorf("expected cpu=500m, got %q", c.Limits.CPU)
	}
	if !c.Probes.HasReadiness {
		t.Error("expected readiness probe")
	}
	if !c.Probes.HasLiveness {
		t.Error("expected liveness probe")
	}
}

func TestParseWorkload_NoProbes(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "deploy.yaml")
	os.WriteFile(f, []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: no-probes
spec:
  template:
    spec:
      containers:
        - name: app
          image: app:latest
`), 0644)

	w, err := ParseWorkload(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c := w.Containers[0]
	if c.Probes.HasReadiness {
		t.Error("expected no readiness probe")
	}
	if c.Probes.HasLiveness {
		t.Error("expected no liveness probe")
	}
	if len(c.Env) != 0 {
		t.Errorf("expected empty env, got %v", c.Env)
	}
}

func TestParseWorkload_NoResources(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "deploy.yaml")
	os.WriteFile(f, []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: no-resources
spec:
  template:
    spec:
      containers:
        - name: app
          image: app:latest
`), 0644)

	w, _ := ParseWorkload(f)
	c := w.Containers[0]
	if c.Limits.Raw != "" {
		t.Error("expected no memory limit")
	}
}

func TestParseWorkload_NotAWorkload(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "svc.yaml")
	os.WriteFile(f, []byte("apiVersion: v1\nkind: Service\n"), 0644)

	w, err := ParseWorkload(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w != nil {
		t.Error("expected nil for non-workload")
	}
}

func TestParseService_LoadBalancer(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "svc.yaml")
	os.WriteFile(f, []byte(`apiVersion: v1
kind: Service
metadata:
  name: otel-collector
spec:
  type: LoadBalancer
`), 0644)

	svc, err := ParseService(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc.Type != "LoadBalancer" {
		t.Errorf("expected LoadBalancer, got %q", svc.Type)
	}
	if svc.Name != "otel-collector" {
		t.Errorf("expected otel-collector, got %q", svc.Name)
	}
}

func TestParseService_DefaultClusterIP(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "svc.yaml")
	os.WriteFile(f, []byte("apiVersion: v1\nkind: Service\nmetadata:\n  name: internal\n"), 0644)

	svc, _ := ParseService(f)
	if svc.Type != "ClusterIP" {
		t.Errorf("expected default ClusterIP, got %q", svc.Type)
	}
}

func TestParseMemoryMiB(t *testing.T) {
	tests := []struct {
		input string
		want  int64
		ok    bool
	}{
		{"512Mi", 512, true},
		{"1Gi", 1024, true},
		{"256M", 244, true},
		{"1G", 953, true},
		{"128Ki", 0, true},
		{"", 0, false},
		{"abc", 0, false},
	}

	for _, tt := range tests {
		got, ok := parseMemoryMiB(tt.input)
		if ok != tt.ok {
			t.Errorf("parseMemoryMiB(%q) ok=%v, want %v", tt.input, ok, tt.ok)
		}
		if ok && got != tt.want {
			t.Errorf("parseMemoryMiB(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}
