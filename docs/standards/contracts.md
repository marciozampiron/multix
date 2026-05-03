# Multix v1.0 Public Contracts

This document captures the **stability promise** Multix makes to anyone
extending the runtime — adapter authors, plugin developers, AI-agent
integrators and library consumers.

It is not aspirational; every claim here is enforced by code (interface
assertions + contract tests under `test/skill_contract_test.go` and the
package `_test.go` files under each adapter).

## What is "v1.0 public surface"?

Three packages form the entire stable extension surface:

| Package                      | Role                                                          |
|------------------------------|---------------------------------------------------------------|
| `internal/domain/skills`     | The universal `Skill` interface every capability implements   |
| `internal/ports/outbound`    | Cloud / AI / k8s provider contracts adapters must satisfy     |
| `internal/ports/inbound`     | CLI handler and AI agent tool-calling contracts               |

> Note on `internal/`: today these packages live under `internal/` because
> the project hasn't carved out a public Go module path yet. Once that
> happens, they will migrate to a stable import path. The contract semantics
> are already v1.0; only the import path is in flux. Plugin implementation is
> tracked separately under issue #20 and ADR 0008.

## Semver promise (v1.0+)

For every interface in the three packages above:

1. **Method signatures are frozen.** Renaming a method, adding a parameter,
   removing a return value, or changing semantic behaviour requires a major
   version bump (`/v2`, `/v3`).

2. **New methods MAY be added in a minor release** *only* via the
   "embedded Plus" pattern. Example:

   ```go
   // Stable since v1.0
   type AuthProvider interface { Login(...); Whoami(...); Validate(...) }

   // Added in v1.4 — does NOT break v1.0 implementations
   type AuthProviderPlus interface {
       AuthProvider
       Refresh(ctx context.Context) error
   }
   ```

   Code that needs the new method does a runtime type-assert; code that
   only knows v1.0 keeps compiling.

3. **Method-comment behaviour is part of the contract.** Adapters that
   silently violate documented invariants (e.g. returning a stale cached
   identity from `Whoami`) are non-conformant even if they compile.

4. **Skill names are forever.** Once `inventory.scan` ships, that exact
   name is reserved for that exact capability. Repurposing it is a major
   version bump of the skill.

## Adapter author checklist

An adapter is "v1-conformant" when:

- [ ] It declares compile-time interface compliance:
      `var _ outbound.AuthProvider = (*Adapter)(nil)`
- [ ] It is safe for concurrent use (the registry shares one instance).
- [ ] Errors wrap the underlying SDK error with `%w` and include a
      remediation hint (env var, command, doc link).
- [ ] It ships a contract test under its own package
      (see `internal/adapters/outbound/cloud/aws/contract_test.go`).
- [ ] CHANGELOG mentions the adapter and the SDK version it pins.

## Skill author checklist

A skill is "v1-conformant" when:

- [ ] `Name()` matches `^[a-z][a-z0-9_]*\.[a-z][a-z0-9_]*$`
      (`<domain>.<verb>`, lowercase, snake-case allowed within tokens).
- [ ] `Description()` reads as a sentence an LLM can use to decide *when*
      to call the skill (not what code it executes).
- [ ] `InputSchema()` returns a JSON-Schema-shaped object literal whose
      top-level `type` is `"object"`.
- [ ] `Execute()` is safe for concurrent invocation.
- [ ] The skill is registered exactly once in `internal/bootstrap/skills.go`.

The runtime contract test (`TestSkillContract_AllRegisteredSkills`)
mechanically enforces the first three items.

## Outside the public surface

Anything under `internal/adapters/`, `internal/application/`,
`internal/bootstrap/`, `internal/platform/`, `internal/domain/<not-skills>`
and `cmd/` is **internal**. It can be refactored or replaced without notice
between minor versions. Do not import these from a plugin or external
module.

## Contract violations: how to report

Open an issue with the label `area/contracts`. The maintainer rotation
treats contract regressions as P0 — a regression that ships triggers a
patch release within 48h.
