# Roadmap

## v0.1.0 — Initial Release

- [x] CLI skeleton (analyze, version, fix, graph, explain)
- [x] YAML parser with line/column source locations
- [x] Pipeline graph engine
- [x] Text, JSON, SARIF output formatters
- [x] Rule engine with 38 rules
- [x] Policy file support (.oteldoctor.yaml)
- [x] Safe autofix (OTEL-REL-102: reorder memory_limiter)
- [x] Directory scanning with glob filters
- [x] Kubernetes ConfigMap extraction
- [x] Kubernetes readiness rules
- [x] GitHub Action
- [x] GoReleaser cross-platform builds

## v0.2.0 — Planned

- [ ] Helm chart value extraction (`values.yaml` → collector config)
- [ ] Rule severity customization via config file annotations
- [ ] Multi-file/cross-document analysis (e.g., receiver in one file, pipeline in another)
- [ ] Kubernetes workload-to-ConfigMap linking via volume/envFrom references
- [ ] Performance optimization for large directories (100+ YAML files)
- [ ] oteldoctor docker image

## v0.3.0 — Planned

- [ ] Custom rule plugins (WebAssembly or Lua)
- [ ] IDE integration (VS Code extension)
- [ ] Pre-commit hook
- [ ] OpenTelemetry Protocol (OTLP) collector config validation via live collector

## Ideas

- Annotate YAML with inline suppression comments (`# oteldoctor: ignore OTEL-REL-103`)
- Compare two configs and show semantic differences
- Generate a collector config from a high-level description
