# oteldoctor

A static analysis tool for OpenTelemetry Collector configuration files.

oteldoctor parses OpenTelemetry Collector YAML configurations and reports:

- **Structural issues** — invalid YAML, missing required keys, type mismatches
- **Reliability problems** — retry/queue misconfiguration, missing exporters
- **Security vulnerabilities** — exposed endpoints, weak TLS, missing auth
- **Cost/cardinality risks** — unbounded attributes, high-cardinality span names
- **Semantic convention violations** — non-standard resource attributes
- **Kubernetes readiness gaps** — missing probes, resource limits

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

# Print version
oteldoctor version
```

## Development

```bash
make build   # build binary
make test    # run tests
make lint    # run linter (TBD)
make clean   # remove build artifacts
```
