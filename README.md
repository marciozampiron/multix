# MULTIX — AI-Native Multi-Cloud Skill Runtime

## 1. Title + Tagline

Enterprise-grade multi-cloud operations for humans, command-line workflows, and AI agents.

## 2. What MULTIX Is

MULTIX is not just a CLI. It is a reusable execution runtime for operational skills that can be invoked consistently across multiple surfaces:
- CLI commands
- AI agents and function-calling adapters
- Future local runtime endpoints (`multix serve`)
- Future APIs and automation layers

Core principle:
> **Agent thinks and orchestrates. Skill executes.**

**Implemented and validated today:**
- Go project with Cobra CLI
- Skills-first architecture and universal skill contract
- Skill registry and skill executor
- Provider registry with normalized provider names
- Agent tool adapter and tool manifest generation
- Real AWS auth skills: `auth.validate`, `auth.whoami`
- Real GCP auth skills: `auth.validate`, `auth.whoami`
- GCP whoami enrichment: ADC first, env fallback, optional local `gcloud` enrichment
- Tests, build, vet, and fmt passing locally

**Implemented target (v0.4-alpha):**
- OCI auth provider (preview)
- `multix serve` local runtime
- Capability matrix endpoint
- Agent runtime endpoints (health, manifests, tool execution)

**Planned / next target (v0.5-alpha):**
- Inventory real providers for AWS, GCP, OCI

## 3. Architectural Philosophy

Traditional CLIs are provider-endpoint-centric. MULTIX is capability-centric and built around **reusable skills** that are stable across providers and execution surfaces.

Examples of skills (some implemented, some foundational/planned):
- `auth.validate`
- `auth.whoami`
- `inventory.scan`
- `inventory.summary`
- `k8s.list_clusters`
- `ai.explain`

The CLI is only one surface. Skills are the product.

## 4. Universal Skill Contract

Skills implement a JSON-capable contract and can be executed consistently across surfaces.

- CLI handlers translate flags into a JSON payload
- The Skill Executor runs the capability
- Agent tooling consumes manifests generated from the same registry

**Register once. Expose everywhere.**

## Runtime (v0.4-alpha)

`multix serve` exposes the local HTTP runtime

Runtime endpoints:
- `GET /health`
- `GET /tools`
- `POST /execute`
- `GET /capabilities`

Links:
- [docs/runtime-architecture.md](docs/runtime-architecture.md)
- [docs/quickstart-agent-runtime.md](docs/quickstart-agent-runtime.md)

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

**Prerequisites:** Go 1.25+

```bash
go mod tidy
make build
./build/multix version
```

## 7. Current CLI Usage

Common global flags:
- `--provider`
- `--output [json|table]`

Provider values currently vary by skill:
- Cloud providers: `aws`, `gcp`
- AI-specific provider values: `gemini`

Real, validated auth commands:

```bash
./build/multix auth validate --provider aws --output json
./build/multix auth whoami --provider aws --output table
./build/multix auth validate --provider gcp --output json
./build/multix auth whoami --provider gcp --output table
```

## 8. Local Auth Troubleshooting

### AWS notes
- AWS SDK for Go v2 requires region and credentials
- Recommended environment for local development:
- `export AWS_REGION=us-east-1`
- `export AWS_EC2_METADATA_DISABLED=true`
- `export AWS_PROFILE=<profile>`

### GCP notes
- `gcloud auth login` authenticates the gcloud CLI
- `gcloud auth application-default login` authenticates Application Default Credentials for SDKs/apps
- These are related but distinct
- Recommended local setup:
- `gcloud auth login`
- `gcloud auth application-default login`
- `gcloud config set project <project-id>`

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

### v0.4-alpha — OCI Auth + Agent Runtime (`multix serve`)
- OCI auth provider (preview)
- `multix serve`
- Health endpoint
- Manifests endpoint
- Tool execution endpoint
- Capability matrix endpoint
- Stronger positioning as a runtime for humans and agents

## 10. Forward Roadmap

### v0.5-alpha — Real Inventory
- AWS: EC2 + S3
- GCP: Compute + Cloud Storage
- OCI: Compute + Object Storage

### v0.6-alpha — Kubernetes Real Integrations
- EKS
- GKE
- OKE

### v0.7-alpha — Golden Paths / Operational Skills
- `doctor.auth`
- `doctor.k8s`
- `k8s.cluster-health`
- `landingzone.audit`
- `security.identity-posture`
- `cost.quick-scan`

### v1.0.0-beta — Enterprise Hardening
- Stable provider contracts
- Stable skill contracts
- Hardened runtime
- Plugin story / extension model
- Enterprise docs + skill catalog

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
- Future-ready for MCP-style tool ecosystems

In practical terms, MULTIX is evolving from a CLI into a local execution runtime for portable, provider-aware operational skills.

## 13. License

Define your preferred license here (Apache-2.0, MIT, or internal enterprise license).
