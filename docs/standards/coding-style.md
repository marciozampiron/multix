# Coding Style Overview

## Principles

- Idiomatic Go first
- Skills-first architecture first
- Small interfaces
- Explicit wiring
- Thin CLI handlers
- Reusable skills
- Fast startup
- Low dependency footprint

---

## Always Prefer

- `gofmt`
- `context.Context` in I/O
- `log/slog`
- table-driven tests where useful
- minimal diffs
- provider isolation behind ports
- clear naming over clever naming

---

## Avoid

- unnecessary abstractions
- heavy DDD ceremony
- giant service layers
- provider-first organization
- business logic in Cobra handlers
- noisy comments
- large refactors without reason