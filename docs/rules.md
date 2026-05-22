# Rules Reference

oteldoctor ships with 38 rules across 6 categories. Use `oteldoctor explain <RULE_ID>` for detailed documentation including bad/good examples and fix guidance.

## Structural (8 rules)

| Rule ID | Title | Severity |
|---|---|---|
| OTEL-STRUCT-001 | Undefined receiver reference | critical |
| OTEL-STRUCT-002 | Undefined processor reference | critical |
| OTEL-STRUCT-003 | Undefined exporter reference | critical |
| OTEL-STRUCT-004 | Undefined extension reference | high |
| OTEL-STRUCT-005 | Unused receiver | low |
| OTEL-STRUCT-006 | Unused processor | low |
| OTEL-STRUCT-007 | Unused exporter | low |
| OTEL-STRUCT-008 | Empty pipeline | high |

## Reliability (8 rules)

| Rule ID | Title | Severity |
|---|---|---|
| OTEL-REL-101 | memory_limiter processor missing | high (production), medium (staging) |
| OTEL-REL-102 | memory_limiter not first processor | high (prod), medium (staging), low (dev) |
| OTEL-REL-103 | Batch processor missing | medium (prod), low (dev) |
| OTEL-REL-104 | Batch before transform/filter | medium (prod), low (dev) |
| OTEL-REL-105 | Exporter retry_on_failure not configured | medium (prod), low (staging/dev) |
| OTEL-REL-106 | Exporter sending_queue not configured | medium (prod), low (staging/dev) |
| OTEL-REL-107 | health_check extension missing in production | medium |
| OTEL-REL-108 | health_check extension not enabled | low |

## Security (7 rules)

| Rule ID | Title | Severity |
|---|---|---|
| OTEL-SEC-201 | Plain HTTP exporter endpoint in production | high |
| OTEL-SEC-202 | Possible hardcoded secret | critical |
| OTEL-SEC-203 | Receiver bound to 0.0.0.0 without auth/TLS | high (production only) |
| OTEL-SEC-204 | Debug exporter used in production | medium |
| OTEL-SEC-205 | Pprof or zpages on public interface | high (prod), medium (staging), low (dev) |
| OTEL-SEC-206 | TLS configuration missing for remote exporter | high (prod), medium (staging/dev) |
| OTEL-SEC-207 | Auth extension not enabled | medium |

## Cost / Cardinality (5 rules)

| Rule ID | Title | Severity |
|---|---|---|
| OTEL-COST-301 | High-cardinality attribute in metric dimensions | medium |
| OTEL-COST-302 | Transform processor creates dynamic attributes | medium |
| OTEL-COST-303 | Spanmetrics connector uses risky dimensions | medium |
| OTEL-COST-304 | Debug exporter may emit high volume | medium (prod/staging) |
| OTEL-COST-305 | Traces pipeline has no sampling | high (production only) |

## Semantic Quality (5 rules)

| Rule ID | Title | Severity |
|---|---|---|
| OTEL-SEM-401 | service.name not configured | low |
| OTEL-SEM-402 | Legacy service name attribute used | medium |
| OTEL-SEM-403 | deployment.environment missing in production | medium |
| OTEL-SEM-404 | service.version missing in production | low |
| OTEL-SEM-405 | Deprecated semantic attribute detected | low |

## Kubernetes Readiness (5 rules)

| Rule ID | Title | Severity |
|---|---|---|
| OTEL-K8S-501 | memory_limiter too close to container limit | medium |
| OTEL-K8S-502 | GOMEMLIMIT missing | medium |
| OTEL-K8S-503 | Container resources not defined | medium |
| OTEL-K8S-504 | Kubernetes probe missing for health_check | medium |
| OTEL-K8S-505 | Receiver exposed as LoadBalancer | medium (production only) |

## Profile-Dependent Severity

Many rules adjust severity based on the analysis profile:

- **development**: Lower severity, some rules disabled entirely (SEC-203, SEC-205, SEC-206, COST-304, COST-305, K8S-505)
- **staging**: Medium severity, reliability/security rules partially enabled
- **production**: Highest severity, all rules enabled

Use `--profile production` for deployment-ready configs and `--profile development` for local iteration.
