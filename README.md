# oteldoctor

A static analysis tool for OpenTelemetry Collector configuration files.

oteldoctor parses OpenTelemetry Collector YAML configurations and reports:

- **Structural issues** — undefined references, unused components, empty pipelines
- **Reliability problems** — missing memory_limiter/batch, retry/queue misconfiguration
- **Security vulnerabilities** — plain HTTP, hardcoded secrets, exposed endpoints, missing TLS
- **Cost/cardinality risks** — high-cardinality attributes, missing sampling, debug in production
- **Semantic convention violations** — deprecated attributes, missing service.name

## Installation

```bash
go install github.com/firfircelik/oteldoctor/cmd/oteldoctor@latest
```

Or build from source:

```bash
make build
./bin/oteldoctor --help
```

## Usage

```bash
# Analyze a single file
oteldoctor analyze config.yaml

# Analyze with JSON output
oteldoctor analyze config.yaml --format json

# Analyze a directory (scans .yaml/.yml recursively)
oteldoctor analyze ./configs/

# Production profile with stricter checks
oteldoctor analyze --profile production config.yaml

# Show suppressed diagnostics
oteldoctor analyze --policy .oteldoctor.yaml --show-suppressed config.yaml

# Auto-fix safe issues (dry-run by default)
oteldoctor fix config.yaml
oteldoctor fix --write config.yaml

# Print version
oteldoctor version
```

## Example Configurations

See `examples/` for reference configurations:

```
examples/
  good/
    collector-dev.yaml          # Clean dev config (0 issues)
    collector-production.yaml   # Clean production config (0 issues)
    semantic.yaml               # Clean semantic conventions
  bad/
    structural.yaml              # Undefined refs, unused, empty pipelines
    reliability.yaml            # Missing memory_limiter, batch order, retry/queue
    security.yaml               # Plain HTTP, hardcoded secret, exposed pprof
    cost.yaml                   # High-cardinality attrs, dynamic attributes, no sampling
    semantic.yaml               # Legacy service name, deprecated attributes
  k8s/
    configmap.yaml              # K8s ConfigMap with embedded collector config
    deployment.yaml             # K8s Deployment manifest
```

## Development

```bash
make build   # build binary
make test    # run tests
make lint    # run linter (TBD)
make clean   # remove build artifacts
```
