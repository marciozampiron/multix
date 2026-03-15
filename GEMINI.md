# GEMINI.md — Multix Repository Instructions

You are working inside the **Multix** repository.

Multix is a **skills-first, agent-ready, multi-cloud CLI** written in Go, focused on:
- DevOps
- Cloud Engineering
- Kubernetes
- AI-assisted workflows
- reusable skills for future AI agents

This repository must remain:
- idiomatic Go
- modular
- fast to start
- low-overhead
- easy to evolve
- open-source product quality

---

## 1. Core Product Vision

Multix is not just a CLI.
It is a **platform of reusable capabilities ("skills")** that can be consumed by:
- CLI commands
- future API endpoints
- future AI agents / tool-calling
- future MCP-compatible adapters

### Golden Rule
- **Agent thinks**
- **Skill executes**

The agent must never own the business logic.
Business logic belongs to:
- domain
- application/use cases
- reusable skills

---

## 2. Mandatory Architectural Rules

### 2.1 Skills-first architecture
Organize the repository by **SKILLS / CAPABILITIES**, not by cloud provider.

Correct:
- `doctor`
- `auth`
- `inventory`
- `k8s`
- `ai`
- `plugin`

Wrong:
- organizing the entire project around `aws/`, `gcp/`, `azure/` as the primary axis

Cloud providers and AI providers must live behind:
- ports/interfaces
- adapters
- explicit wiring

### 2.2 Layering rules
Use this mental model:

CLI / API / Agent
→ Skill / Capability / Tool
→ Application Use Case
→ Port (interface)
→ Adapter / Provider
→ External system (AWS, GCP, Gemini, etc.)

### 2.3 Go layout rules
Use idiomatic Go repository structure:
- `cmd/` for executable entrypoints
- `internal/` for non-exported implementation
- `internal/application/` for use cases
- `internal/ports/` for contracts
- `internal/adapters/` for inbound/outbound integrations
- `internal/bootstrap/` for explicit wiring
- avoid `pkg/` unless there is a strong reason

---

## 3. Controlled Change Rule (Mandatory)

For non-trivial repository upgrades:
- audit first
- propose a minimal change plan
- prefer patch-style evolution
- avoid redesign unless explicitly requested
- implementation should happen only after plan approval

For upgrade tasks:
1. inspect the current repository
2. map what already exists
3. identify the smallest possible delta
4. prefer modifying existing files over creating new files
5. avoid moving files unless strictly necessary
6. avoid renaming packages unless strictly necessary
7. avoid duplicate abstractions
8. preserve package names and import paths whenever possible

---

## 4. Non-Negotiable Engineering Principles

- Prefer **idiomatic Go** over theoretical purity
- Prefer **clarity** over cleverness
- Prefer **explicit wiring** over magic
- Prefer **small interfaces** near consumers
- Prefer **composition** over inheritance
- Prefer **pragmatic DDD Lite** over heavy enterprise DDD
- Prefer **simple, stable contracts** over over-abstraction
- Prefer **skills reusable by CLI/API/Agent** over agent-only logic

---

## 5. What You MUST Avoid

Do NOT:
- create unnecessary interfaces
- create service layers without real value
- create manager/service/repository abstractions by habit
- over-engineer the MVP
- put business logic inside Cobra command handlers
- make the architecture provider-first
- initialize all providers eagerly on startup
- use reflection unless clearly justified
- use native Go `.so` plugin loading in MVP
- hardcode secrets or credentials
- log secrets
- introduce large dependencies without justification
- create duplicate registries
- create duplicate skill contracts
- create duplicate command trees
- create parallel bootstrap paths

---

## 6. Skill Model (Very Important)

A skill is a reusable capability of the platform with a stable input/output contract.

A skill must be callable by:
- CLI
- future API
- future AI agents

Each skill should be:
- small
- composable
- machine-friendly
- deterministic when possible
- easy to test
- provider-agnostic when possible

Naming convention:
- `<domain>.<action>`

Examples:
- `doctor.run`
- `auth.validate`
- `auth.whoami`
- `inventory.scan`
- `inventory.summary`
- `k8s.list_clusters`
- `ai.explain`
- `ai.generate_terraform`

---

## 7. Documentation-in-Code Rules (Mandatory)

### 7.1 File headers (mandatory for important files)
Important source files must include a short standardized file header at the top.

Recommended format:

```go
// File: internal/application/inventory/scan.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Purpose: Executes the inventory scan use case for reusable platform and agent skill flows.
```

---

## 8. Documentation-Only Pass Rule

When asked to improve comments or documentation in the repository:
- prefer a documentation-only pass
- do not change business logic
- do not refactor behavior
- do not rename packages or symbols
- do not move files
- do not introduce new abstractions

Documentation-only changes may include:
- file headers in important Go files
- Go doc comments for exported symbols
- concise rationale comments for non-obvious code

Avoid:
- noisy comments
- comments that describe obvious mechanics
- over-commenting idiomatic Go code