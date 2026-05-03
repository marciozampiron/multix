# ADR 0004 — Local HTTP Runtime (`multix serve`)

## Status
Accepted

## Context
Skills must be reachable from three distinct callers — the human at the
terminal, an external AI agent loop, and a future MCP server. A common
machine-readable surface is required so skills are written once and
exposed everywhere without per-caller adapters.

Two alternatives were considered:

- **In-process Go API only.** Forces every consumer to be Go and to
  embed the binary. Rules out remote agents and language interop.
- **gRPC service.** Heavy dependency footprint; tooling is harder for
  curl-driven smoke tests and quick LLM iteration.

## Decision
Ship a local HTTP runtime via `multix serve`, exposing four endpoints:

- `GET /health` — liveness probe.
- `GET /tools` — manifest of every registered skill (LLM tool list).
- `GET /capabilities` — runtime version + supported providers + feature flags.
- `POST /execute` — `{skill, provider, params}` invocation.

JSON in/out, port configurable, no auth in v1 (loopback by default).

## Consequences
- Any language can drive Multix via stdlib HTTP + JSON.
- Smoke tests use `curl`; agents use HTTP libraries; MCP gets a thin
  shim that bridges JSON-RPC → `/execute`.
- Security model is "trust the loopback"; multi-tenant deployment will
  require a reverse proxy with auth (deferred to v1.5).
- Prometheus metrics (`/metrics`, ADR 0006) and request-ID middleware
  (ADR 0005) bolt onto this same mux.
