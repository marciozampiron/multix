
---

# 2) `.agent/ARCHITECTURE.md`

```md
# Multix Architecture Rules

## 1. Architecture Style

Multix uses:
- Skills-first architecture
- DDD Lite (pragmatic)
- Clean / Hexagonal architecture (pragmatic)
- Ports and Adapters
- explicit bootstrap / wiring

This is NOT heavy DDD.
This is NOT enterprise Java style layering.

The design goal is:
- modularity
- testability
- performance
- extensibility
- low cognitive overhead

---

## 2. Primary Axis of Organization

The repository is organized around **capabilities (skills)**.

Primary domains:
- doctor
- auth
- inventory
- k8s
- ai
- plugin

Future domains:
- cost
- security
- terraform
- drift
- policy
- finops
- compliance
- remediation

Do NOT reorganize the repo around cloud providers.

---

## 3. Layer Responsibilities

### 3.1 `cmd/`
Contains executable entrypoints only.

No business logic here.

### 3.2 `internal/application/`
Contains use cases / orchestration logic.

Responsibilities:
- call ports
- coordinate flows
- transform domain data for use case results
- remain independent from Cobra and concrete SDKs

### 3.3 `internal/ports/`
Contains contracts/interfaces.

Rules:
- interfaces should be small
- interfaces should be consumer-oriented
- do not create giant god-interfaces

### 3.4 `internal/adapters/inbound/`
Contains entrypoint adapters.

Examples:
- Cobra command handlers
- future API handlers
- future agent tool adapters

Rules:
- no business logic here
- parse inputs
- call skill or use case
- format outputs

### 3.5 `internal/adapters/outbound/`
Contains integrations with external systems.

Examples:
- AWS adapter
- GCP adapter
- Gemini adapter

Rules:
- isolate provider-specific SDK details here
- do not leak SDK types into application layer

### 3.6 `internal/bootstrap/`
Contains explicit wiring.

Responsibilities:
- construct providers
- construct registries
- wire use cases
- register skills
- build root command

Rules:
- no DI framework
- explicit construction
- easy to read
- easy to test

### 3.7 `internal/domain/`
Contains:
- models
- domain-specific errors
- skill definitions / metadata
- shared domain semantics

Keep it lean.

---

## 4. Skill Design Rules

A skill is a reusable executable capability.

A skill should:
- map to a meaningful capability
- be reusable by CLI/API/Agent
- be testable
- have stable input/output keys
- avoid provider lock-in when possible

---

## 5. Provider Design Rules

Providers should be:
- behind ports
- lazily selected when possible
- isolated from core packages
- replaceable

Do not:
- reference concrete provider SDK types in application layer
- pass SDK structs through ports

---

## 6. Plugin System Rules

MVP plugin design should be:
- manifest-based
- registry-based
- explicit
- safe

Do NOT use:
- native Go `.so` plugin loading in MVP

---

## 7. Future Agent Layer

Future agent adapters should call skills, not re-implement logic.

Rule:
- agents orchestrate
- skills execute