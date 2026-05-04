# MULTIX — AI-Native Multi-Cloud Skill Runtime

[![CI](https://github.com/marciozampiron/multix/actions/workflows/ci.yml/badge.svg)](https://github.com/marciozampiron/multix/actions/workflows/ci.yml)

## 1. Title + Tagline

Enterprise-grade multi-cloud operations for humans, command-line workflows, and AI agents.

## 2. What MULTIX Is

MULTIX is not just a CLI. It is a reusable execution runtime for operational skills that can be invoked consistently across multiple surfaces:
- CLI commands
- AI agents and function-calling adapters
- Local runtime endpoints (`multix serve`)
- Future MCP/plugin integrations

Core principle:
> **Agent thinks and orchestrates. Skill executes.**

**Implemented and validated today:**
- Go project with Cobra CLI and local HTTP runtime
- Skills-first architecture and universal skill contract
- Stable v1.0 contract docs for skills and provider ports
- Skill registry, skill executor, and dynamic tool manifests
- Provider registry with normalized provider names
- Runtime endpoints for health, tools, capabilities, metrics, and execution
- Structured request IDs propagated through runtime execution
- Real AWS and GCP auth skills; OCI auth adapter support
- Inventory skills for AWS, GCP, and OCI adapters/stubs
- Kubernetes discovery through `k8s.list_clusters`
- FOCUS-aligned `cost.focus_report` with AWS Cost Explorer, GCP BigQuery billing export, and OCI Usage API
- AI-assisted skills for network generation, Kubernetes audit, and IAM audit
- CI and local validation gates for fmt, tests, vet, build, cross-build, and race tests

**Still open:**
- Issue #20: implementation of the plugin extension model proposed in ADR 0008.
- Future MCP server/client surfaces for federated plugins.

## 3. Architectural Philosophy

Traditional CLIs are provider-endpoint-centric. MULTIX is capability-centric and built around **reusable skills** that are stable across providers and execution surfaces.

Examples of skills (some implemented, some foundational/planned):
- `auth.validate`
- `auth.whoami`
- `inventory.scan`
- `inventory.summary`
- `k8s.list_clusters`
- `doctor.auth`
- `doctor.k8s`
- `ai.explain`
- `cost.quick_scan`
- `cost.focus_report`
- `infra.generate_network`
- `landingzone.audit`
- `security.identity_posture`
- `security.k8s_audit`
- `security.iam_audit`

The CLI is only one surface. Skills are the product.

## 4. Universal Skill Contract

Skills implement a JSON-capable contract and can be executed consistently across surfaces.

- CLI handlers translate flags into a JSON payload
- The Skill Executor runs the capability
- Agent tooling consumes manifests generated from the same registry

**Register once. Expose everywhere.**

## Runtime

`multix serve` exposes the local HTTP runtime

Runtime endpoints:
- `GET /health`
- `GET /tools`
- `POST /execute`
- `GET /capabilities`
- `GET /metrics`

Links:
- [docs/runtime-architecture.md](docs/runtime-architecture.md)
- [docs/quickstart-agent-runtime.md](docs/quickstart-agent-runtime.md)
- [docs/testing-playbook.md](docs/testing-playbook.md)

## 5. Project Layout

```text
multix/
├── GEMINI.md
├── .agent/
├── cmd/
├── docs/
├── internal/
├── pkg/
├── prompts/
├── test/
├── Makefile
├── README.md
├── go.mod
└── go.sum
```

## 6. Quick Start

**Prerequisites:** Go 1.25+ or the version declared in `go.mod`.

```bash
go mod tidy
make build
./build/multix version
```

Full local gate:

```bash
make all
make build-cross
make test-race
git diff --check
```

## 7. Command Reference

Build once before running local examples:

```bash
make build
```

Global flags:

- `--provider`, `-p`: target provider. Default: `aws`.
- `--output`, `-o`: output format. Supported values: `json`, `table`.
- `--log-level`, `-l`: log level. Supported values: `debug`, `info`, `warn`, `error`.

Provider values:

- Cloud providers: `aws`, `gcp`, `oci`.
- AI provider: `gemini` is used by `ai.explain`.

### Version

```bash
./build/multix version
```

Prints the current CLI version compiled into the binary.

### Doctor

```bash
./build/multix doctor
```

Runs local readiness checks for tools and dependencies through the `doctor.run` skill.

Runtime-only related skills:

- `doctor.auth`: provider auth health summary.
- `doctor.k8s`: managed Kubernetes health summary.

Use them through `/execute`:

```bash
curl -s -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{"skill":"doctor.auth","params":{"providers":["aws","gcp","oci"]}}'
```

```bash
curl -s -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{"skill":"doctor.k8s","params":{"providers":["aws","gcp","oci"]}}'
```

### Auth

```bash
./build/multix auth validate --provider aws --output json
./build/multix auth validate --provider gcp --output json
./build/multix auth validate --provider oci --output json
```

Validates credentials for the selected provider through `auth.validate`.

```bash
./build/multix auth whoami --provider aws --output table
./build/multix auth whoami --provider gcp --output table
./build/multix auth whoami --provider oci --output json
```

Displays the active identity for the selected provider through `auth.whoami`.

Runtime-only related skill:

- `auth.login`: login-oriented skill exposed through the registry.

```bash
curl -s -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{"skill":"auth.login","provider":"aws","params":{}}'
```

### Inventory

```bash
./build/multix inventory list --provider aws --service compute --output json
./build/multix inventory list --provider aws --service storage --output table
./build/multix inventory list --provider gcp --service compute --output json
./build/multix inventory list --provider gcp --service storage --output json
./build/multix inventory list --provider oci --service compute --output json
./build/multix inventory list --provider oci --service storage --output json
```

Runs `inventory.scan` for a provider and optional service/resource type.

Runtime-only related skill:

- `inventory.summary`: provider inventory summary.

```bash
curl -s -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{"skill":"inventory.summary","params":{"provider":"aws"}}'
```

### Kubernetes

```bash
./build/multix k8s clusters --provider aws
./build/multix k8s clusters --provider gcp
./build/multix k8s clusters --provider oci
```

Lists managed clusters through `k8s.list_clusters`. The current CLI renderer prints a compact table-style line per cluster.

### AI

```bash
./build/multix ai explain "Explain this IAM finding" --output json
```

Runs `ai.explain` using the Gemini provider path.

### Runtime Server

```bash
./build/multix serve --port 8080
```

Starts the local HTTP runtime for agents and automation.

Endpoints:

```bash
curl -s http://localhost:8080/health
curl -s http://localhost:8080/tools
curl -s http://localhost:8080/capabilities
curl -s http://localhost:8080/metrics
```

Execute any registered skill:

```bash
curl -s -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: manual-test-001" \
  -d '{
    "skill": "auth.validate",
    "provider": "aws",
    "params": {}
  }'
```

### Cost And FinOps

`cost.quick_scan` is a fast resource-count proxy:

```bash
curl -s -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "skill": "cost.quick_scan",
    "params": {
      "providers": ["aws", "gcp", "oci"]
    }
  }'
```

`cost.focus_report` returns FOCUS-aligned billing rows:

```bash
curl -s -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "skill": "cost.focus_report",
    "params": {
      "providers": ["aws", "gcp", "oci"],
      "period": "current_month"
    }
  }'
```

Provider prerequisites:

- AWS: Cost Explorer permission for `ce:GetCostAndUsage`.
- GCP: `MULTIX_GCP_BILLING_DATASET=<project>.<dataset>.<table>` and BigQuery billing export access.
- OCI: `MULTIX_OCI_TENANCY_OCID=<tenancy_ocid>` and Usage API access.

Supported periods:

- `current_month`
- `last_30_days`
- `last_month`

### Landing Zone

`landingzone.audit` is exposed through the runtime:

```bash
curl -s -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "skill": "landingzone.audit",
    "params": {
      "providers": ["aws", "gcp", "oci"]
    }
  }'
```

### Security

Identity posture:

```bash
curl -s -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "skill": "security.identity_posture",
    "params": {
      "providers": ["aws", "gcp", "oci"]
    }
  }'
```

Kubernetes audit:

```bash
curl -s -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "skill": "security.k8s_audit",
    "params": {
      "providers": ["aws", "gcp", "oci"],
      "ai_provider": "gemini"
    }
  }'
```

IAM audit:

```bash
curl -s -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "skill": "security.iam_audit",
    "params": {
      "providers": ["aws", "gcp", "oci"],
      "ai_provider": "gemini"
    }
  }'
```

### Infrastructure Generation

`infra.generate_network` generates a network topology proposal through the AI provider:

```bash
curl -s -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "skill": "infra.generate_network",
    "params": {
      "provider": "aws",
      "ai_provider": "gemini",
      "intent": "three-tier web application with public, private, and database subnets",
      "region": "us-east-1",
      "cidr": "10.0.0.0/16"
    }
  }'
```

Use `GET /tools` to inspect the exact input schema for every registered skill:

```bash
curl -s http://localhost:8080/tools
```

## 8. Local Auth Troubleshooting

### AWS notes
- AWS SDK for Go v2 requires region and credentials
- Recommended environment for local development:
- `export AWS_REGION=us-east-1`
- `export AWS_EC2_METADATA_DISABLED=true`
- `export AWS_PROFILE=<profile>`
- Cost reporting uses AWS Cost Explorer and needs permission for `ce:GetCostAndUsage`.

### GCP notes
- `gcloud auth login` authenticates the gcloud CLI
- `gcloud auth application-default login` authenticates Application Default Credentials for SDKs/apps
- These are related but distinct
- Recommended local setup:
- `gcloud auth login`
- `gcloud auth application-default login`
- `gcloud config set project <project-id>`
- GCP billing reports need Cloud Billing BigQuery export and:
- `export MULTIX_GCP_BILLING_DATASET=<project>.<dataset>.<table>`

### OCI notes
- OCI SDK credentials default to `$HOME/.oci/config` unless `OCI_CONFIG_FILE` is set
- Recommended local setup:
- `export OCI_CONFIG_FILE=$HOME/.oci/config`
- `export MULTIX_OCI_TENANCY_OCID=<tenancy_ocid>`
- Cost reporting uses the OCI Usage API and needs Cost and Usage Reports plus usage-budgets read access.

## 9. Product Evolution / Release Journey

### v0.1-alpha — Foundation
- Go module and Cobra bootstrap
- DDD-lite / Hexagonal layout
- Makefile workflow
- Base repository standards

### v0.2-alpha — Skills-First + Agent-Ready
- Universal skill contract
- Skill registry and skill executor
- CLI skill dispatch
- Agent tool adapter
- Tool manifest generation
- Provider abstraction foundations

### v0.3-alpha — Real Auth (AWS + GCP)
- Real AWS `auth.validate` and `auth.whoami`
- Real GCP `auth.validate` and `auth.whoami`
- Provider registry normalization
- Expanded tests

### v0.3.1-alpha — GCP Polish + DX Hardening
- GCP log deduplication
- GCP whoami enrichment (ADC + env + optional `gcloud`)
- Command runner seam for tests
- Local auth troubleshooting docs
- Real local smoke validation

### v0.4 — OCI Auth + Agent Runtime (`multix serve`)
- OCI auth provider (preview)
- `multix serve`
- Health endpoint
- Manifests endpoint
- Tool execution endpoint
- Capability matrix endpoint
- Stronger positioning as a runtime for humans and agents

## 10. Forward Roadmap

### v0.5 — Real Inventory
- AWS: EC2 + S3
- GCP: Compute + Cloud Storage
- OCI: Compute + Object Storage
Status: delivered through provider adapters/stubs and registry wiring.

### v0.6-alpha — Kubernetes Real Integrations
- EKS
- GKE
- OKE
Status: delivered for managed cluster listing paths.

### v0.7-alpha — Golden Paths / Operational Skills
- `doctor.auth`
- `doctor.k8s`
- `doctor.run`
- `landingzone.audit`
- `security.identity_posture`
- `cost.quick_scan`
Status: delivered.

### v0.8-alpha — AI-Assisted Provisioning & Scaffolding
- `infra.generate_network` (VPC/VCN via AI Tooling)
Status: delivered for `infra.generate_network`.

### v0.9-alpha — FinOps Unified Billing (FOCUS)
- `cost.focus_report` (Multi-cloud billing aggregation)
Status: delivered for AWS Cost Explorer, GCP Cloud Billing BigQuery export, and OCI Usage API.

### v0.10-alpha — AI-Assisted Security Remediation
- `security.k8s_audit` (CVE Scanner & AI Remediation)
- `security.iam_audit` (Multi-cloud IAM least-privilege analysis & AI Remediation)
Status: delivered.

### v1.0.0-beta — Enterprise Hardening
- Stable provider contracts
- Stable skill contracts
- Hardened runtime
- Plugin story / extension model
- Enterprise docs + skill catalog
Status: contracts and ADRs delivered; plugin implementation remains tracked under issue #20.

## 11. Extending the Platform

1. Create a skill in `internal/application/<domain>/`.
2. Register it in `internal/bootstrap/skills.go`.
3. Add a CLI handler in `internal/adapters/inbound/cli/`.
4. Agent manifests are generated automatically once the skill is registered.

**Register once. Expose everywhere.**

## 12. Strategic Direction

MULTIX is not just a multi-cloud CLI. It is becoming:
- A universal multi-cloud skill runtime
- A safe execution surface for AI agents
- A foundation for enterprise operational golden paths
- Future-ready for MCP-style plugin ecosystems

In practical terms, MULTIX is evolving from a CLI into a local execution runtime for portable, provider-aware operational skills.

## 13. Testing

Use [docs/testing-playbook.md](docs/testing-playbook.md) as the caderno de teste for local validation, runtime smoke checks, provider checks, and release readiness.

## 14. License

Define your preferred license here (Apache-2.0, MIT, or internal enterprise license).
