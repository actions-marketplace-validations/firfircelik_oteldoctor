# oteldoctor

[![test](https://github.com/firfircelik/oteldoctor/actions/workflows/test.yml/badge.svg)](https://github.com/firfircelik/oteldoctor/actions/workflows/test.yml)
[![Go version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev)
[![license](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A production-readiness analyzer for OpenTelemetry Collector configurations.

## What oteldoctor is

oteldoctor scans OpenTelemetry Collector YAML configurations and reports issues across six categories:

| Category | Examples |
|---|---|
| **Structural** | Undefined references, unused components, empty pipelines |
| **Reliability** | Missing memory_limiter/batch, retry/queue gaps, health_check |
| **Security** | Plain HTTP, hardcoded secrets, exposed debug endpoints, missing TLS |
| **Cost / Cardinality** | High-cardinality metric dimensions, missing sampling, debug in production |
| **Semantic Quality** | Deprecated attributes, missing service.name or deployment.environment |
| **Kubernetes Readiness** | GOMEMLIMIT, resource limits, probes, LoadBalancer exposure |

## Why oteldoctor?

The OpenTelemetry Collector validates whether a config is syntactically valid and can start. oteldoctor goes further — it asks whether your config is production-ready.

- **Before deploying:** catch reliability gaps, security risks, and cost pitfalls that the Collector itself won't flag.
- **In CI:** block merges when critical issues are introduced, with SARIF output for GitHub Code Scanning.
- **For teams:** enforce consistent semantic conventions and Kubernetes best practices across services.

## What oteldoctor is not

oteldoctor does **not** replace the Collector's own configuration validation. It checks production readiness: reliability, security, cost, semantic quality, and Kubernetes deployment best practices.

## Installation

```bash
# Install latest release
go install github.com/firfircelik/oteldoctor/cmd/oteldoctor@v0.1.0

# Or install latest commit
go install github.com/firfircelik/oteldoctor/cmd/oteldoctor@latest
```

Or build from source:

```bash
git clone https://github.com/firfircelik/oteldoctor
cd oteldoctor
make build
./bin/oteldoctor --help
```

## Quick start

```bash
# Analyze a single file
oteldoctor analyze config.yaml

# Analyze a directory (scans .yaml/.yml recursively)
oteldoctor analyze ./deploy/

# Production profile (stricter security and reliability checks)
oteldoctor analyze --profile production ./deploy/

# JSON output (stable for CI diffing)
oteldoctor analyze --format json config.yaml

# SARIF output for GitHub Code Scanning
oteldoctor analyze --format sarif ./deploy/ > oteldoctor.sarif

# Auto-fix safe issues (dry-run by default)
oteldoctor fix config.yaml
oteldoctor fix --write config.yaml

# Render pipeline graph
oteldoctor graph config.yaml --format mermaid

# Explain a rule
oteldoctor explain OTEL-REL-102
```

## Exit codes

| Code | Meaning |
|---|---|
| 0 | No issues found, or all issues below `--fail-on` threshold |
| 1 | Issues found at or above the `--fail-on` threshold |
| 2 | Parse error, config error, or runtime error |

Use `--fail-on critical` to exit 0 unless critical issues exist. Use `--fail-on low` (default) to fail on any issue.

## Example: bad config and output

```yaml
# collector.yaml
receivers:
  otlp: {}
exporters:
  debug: {}
service:
  pipelines:
    traces:
      receivers: [otlp, missing_rcv]
      processors: [undefined_proc]
      exporters: [debug, no_such_exp]
```

```
$ oteldoctor analyze collector.yaml

collector.yaml

CRITICAL OTEL-STRUCT-001 line 7
Pipeline "traces" references receiver "missing_rcv" which is not defined.
Fix: Define a receiver named "missing_rcv" or remove the reference.

CRITICAL OTEL-STRUCT-002 line 7
Pipeline "traces" references processor "undefined_proc" which is not defined.
Fix: Define a processor named "undefined_proc" or remove the reference.

CRITICAL OTEL-STRUCT-003 line 7
Pipeline "traces" references exporter "no_such_exp" which is not defined.
Fix: Define an exporter named "no_such_exp" or remove the reference.

3 issues found.
```

## Rule categories

oteldoctor ships with 38 rules across 6 categories:

```
 8 structural    (OTEL-STRUCT-001 – OTEL-STRUCT-008)
 8 reliability   (OTEL-REL-101  – OTEL-REL-108)
 7 security      (OTEL-SEC-201  – OTEL-SEC-207)
 5 cost          (OTEL-COST-301 – OTEL-COST-305)
 5 semantic      (OTEL-SEM-401  – OTEL-SEM-405)
 5 kubernetes    (OTEL-K8S-501  – OTEL-K8S-505)
```

See [docs/rules.md](docs/rules.md) for the complete rules reference. Use `oteldoctor explain <RULE_ID>` for detailed documentation on any rule.

## Policy file

Create a `.oteldoctor.yaml` in your project root to customize behavior:

```yaml
profile: production
fail_on: high
rules:
  OTEL-REL-105: off          # disable retry_on_failure check
  OTEL-SEC-203: medium       # lower severity for 0.0.0.0 binding
allowed_plain_http_endpoints:
  - http://localhost:8888
suppressions:
  - rule: OTEL-SEC-201
    file: dev-config.yaml
    reason: "local development only"
```

Policy files are discovered automatically (walking up from the current directory) or specified with `--policy path/to/.oteldoctor.yaml`. Suppressed diagnostics are hidden unless `--show-suppressed` is passed.

## GitHub Action

Using the composite action:

```yaml
- uses: firfircelik/oteldoctor@v0.1.0
  with:
    path: ./deploy
    profile: production
    format: sarif
```

Or using `go install` in a workflow step:

```yaml
- uses: actions/setup-go@v5
  with:
    go-version: "1.23"
- run: go install github.com/firfircelik/oteldoctor/cmd/oteldoctor@v0.1.0
- run: oteldoctor analyze --profile production --format sarif ./deploy/ > oteldoctor.sarif
- uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: oteldoctor.sarif
```

See [docs/github-action.md](docs/github-action.md) for full documentation.

## Graph command

```bash
oteldoctor graph config.yaml --format mermaid
```

Produces a Mermaid.js flowchart showing receivers → processors → exporters per signal. Also supports `--format dot` (Graphviz) and `--format json`.

## Explain command

```bash
oteldoctor explain OTEL-REL-102
```

Shows title, category, severity, why the rule matters, bad/good examples, and fix instructions.

## Autofix

```bash
oteldoctor fix config.yaml          # dry-run (default), shows unified diff
oteldoctor fix config.yaml --write  # applies the fix to the file
```

Currently supports one safe fix: **OTEL-REL-102** (reorder memory_limiter to first position in processor chain). Other rules provide fix suggestions in diagnostic output but are not automatically applied.

**Important:** `--write` may change YAML formatting (indentation style, comments). Always review the `--dry-run` diff before writing. For version-controlled configs, commit or stash changes before running `--write`.

## Example configurations

```
examples/
  good/
    collector-dev.yaml           # Clean dev config (0 issues with default profile)
    collector-production.yaml    # Clean production config (0 issues with --profile production)
    semantic.yaml                # Clean semantic conventions
  bad/
    structural.yaml              # Undefined refs, unused components, empty pipelines
    reliability.yaml             # Wrong memory_limiter order, missing retry/queue
    security.yaml                # Plain HTTP, hardcoded secret, exposed pprof, debug in prod
    cost.yaml                    # High-cardinality attrs, dynamic attributes, no sampling
    semantic.yaml                # Legacy app_name, deprecated attributes
  k8s/
    configmap.yaml               # ConfigMap with embedded collector config
    deployment.yaml              # Deployment with probes, resources, GOMEMLIMIT
```

## Development

```bash
make build     # go build -o bin/oteldoctor ./cmd/oteldoctor
make test      # go test -race -count=1 ./...
make clean     # rm -rf bin/
```

See [docs/roadmap.md](docs/roadmap.md) for planned features and [docs/rules.md](docs/rules.md) for the complete rules reference.

## Disclaimer

oteldoctor does **not** validate whether your config will start in the OpenTelemetry Collector. It checks production readiness: reliability risks, security gaps, cost/cardinality exposure, semantic convention quality, and Kubernetes deployment best practices. Always verify your config with the Collector itself before deploying.
