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
The repository must be organized by **SKILLS / CAPABILITIES**, not by cloud provider.

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
Use the following mental model:

CLI / API / Agent
→ Skill / Capability / Tool
→ Application Use Case
→ Port (interface)
→ Adapter / Provider
→ External system (AWS, GCP, Gemini, etc.)

### 2.3 Go layout rules
Follow idiomatic Go repository structure:
- `cmd/` for executable entrypoints
- `internal/` for non-exported implementation
- `internal/application/` for use cases
- `internal/ports/` for contracts
- `internal/adapters/` for inbound/outbound integrations
- `internal/bootstrap/` for explicit wiring
- avoid `pkg/` unless there is a strong, concrete reason

---

## 3. Non-Negotiable Engineering Principles

- Prefer **idiomatic Go** over theoretical purity
- Prefer **clarity** over cleverness
- Prefer **explicit wiring** over magic
- Prefer **small interfaces** near consumers
- Prefer **composition** over inheritance
- Prefer **pragmatic DDD Lite** over heavy enterprise DDD
- Prefer **simple, stable contracts** over over-abstraction
- Prefer **skills reusable by CLI/API/Agent** over agent-only logic

---

## 4. What You MUST Avoid

Do NOT:
- create unnecessary interfaces
- create service layers without real value
- create manager/service/repository abstractions just by habit
- over-engineer the MVP
- put business logic inside Cobra command handlers
- make the architecture provider-first
- initialize all providers eagerly on startup
- use reflection unless clearly justified
- use native Go `.so` plugin loading in MVP
- hardcode secrets or credentials
- log secrets
- introduce large dependencies without justification

---

## 5. Skill Model (Very Important)

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

## 6. Current MVP Scope

### Platform skills
- doctor
- auth
- inventory
- k8s
- ai
- plugin

### Agent-ready skills
- doctor.run
- auth.validate
- auth.whoami
- inventory.scan
- inventory.summary
- k8s.list_clusters
- ai.explain
- ai.generate_terraform

---

## 7. Code Style Rules

Always follow:
- idiomatic Go
- Effective Go
- small functions
- clear names
- minimal nesting
- error-first handling
- `context.Context` in I/O or provider flows
- structured logging
- explicit dependencies

### Comments policy
Do NOT over-comment.
Comments are allowed only when they explain:
- why something exists
- tradeoffs
- non-obvious constraints
- public exported behavior when helpful

Avoid useless comments like:
- "increment counter"
- "call function"
- "set variable"

---

## 8. Performance Rules

Multix is a CLI. Fast startup matters.

Always optimize for:
- fast startup
- low memory overhead
- lazy provider usage
- avoiding unnecessary allocations
- avoiding reflection
- avoiding eager SDK initialization
- keeping central packages lightweight

If a provider SDK is heavy:
- isolate it in outbound adapters
- avoid pulling it into core layers

---

## 9. Security Rules

- Never log secrets
- Never commit credentials
- Keep config separate from secrets
- Mask sensitive data in logs
- Validate inputs
- Prefer least-privilege assumptions
- Prepare for `govulncheck`
- Avoid risky shell execution without explicit reason

---

## 10. Build & Workflow Rules

Prefer the Makefile:
- `make build`
- `make run`
- `make test`
- `make test-race`
- `make fmt`
- `make vet`
- `make vuln`
- `make tidy`

Do not invent random build commands if the Makefile already covers them.

---

## 11. File Change Rules

When making changes:
1. preserve the skills-first architecture
2. avoid unrelated refactors
3. keep diffs minimal and coherent
4. do not rename packages without strong reason
5. if you add a skill:
   - add or reuse use case
   - register it
   - expose it if needed
   - add test(s)
6. if you add a provider:
   - implement the correct port
   - isolate SDK details in adapter
   - avoid leaking provider details into application layer

---

## 12. Preferred Output Behavior for AI Assistance

When asked to modify the codebase:
- first understand the current structure
- then propose the smallest coherent change
- preserve architectural integrity
- explain tradeoffs briefly
- do not rewrite the whole repo unless explicitly asked

When asked to add a feature:
- map it to:
  - domain
  - application use case
  - port (if needed)
  - adapter (if needed)
  - skill (if needed)
  - CLI command (if needed)

---

## 13. Additional Context Files

Read these files as additional project instructions:

@./.agent/ARCHITECTURE.md
@./.agent/SKILL.md
@./.agent/CODING_STANDARDS.md
@./.agent/COMMANDS.md
@./.agent/TESTING.md
@./.agent/SAFETY.md