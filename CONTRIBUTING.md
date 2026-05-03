# Contributing to Multix

Thanks for the interest. Multix is a small project with a tight scope —
keeping it that way is part of the contract.

## TL;DR

1. Open or pick an existing issue before writing code.
2. Branch off `main`. Use `feature/<issue-or-slug>`.
3. Run `make all` locally — fmt + vet + tidy + test + build.
4. Open a PR with `Closes #N` in the body.
5. PRs squash-merge into `main`. One commit per PR.

## Project shape

- `cmd/multix` — binary entrypoint. Stays trivial.
- `internal/domain` — pure domain types (auth, inventory, k8s, skills).
- `internal/ports` — **stable v1 contracts** (see `docs/standards/contracts.md`).
- `internal/application` — skills (the universal capability unit).
- `internal/adapters/inbound` — CLI, runtime HTTP server, agent tool bridge.
- `internal/adapters/outbound` — cloud SDK adapters (AWS, GCP, OCI) + AI providers.
- `internal/bootstrap` — explicit DI wiring.
- `docs/` — runtime architecture, ADRs, standards, skill catalog.

The "Golden Rule" lives in `GEMINI.md`: *agent thinks, skill executes.*
Skills are deterministic; reasoning happens at orchestration time.

## What needs an issue

Anything that touches:
- A package under `internal/ports/` (contract-level changes).
- The runtime HTTP surface (new endpoint, schema change).
- A new skill (must be registered in `internal/bootstrap/skills.go`).
- A new outbound adapter or a new SDK dependency.

Trivial fixes (typo, dead code, dependency bump within a minor) can be
PR-only.

## Dev loop

```bash
git switch -c feature/short-slug main
# ... edit ...
make all          # fmt + vet + tidy + test + build
make test-race    # before opening the PR
```

If you added a skill, the contract test
(`test/skill_contract_test.go`) will exercise it automatically — no
extra work.

If you added an outbound adapter, declare interface compliance:

```go
var _ outbound.AuthProvider = (*Adapter)(nil)
```

and add a `contract_test.go` file under your adapter package.

## Commit style

The repository follows Conventional Commits, lightly:

- `feat(<area>): ...` — new capability
- `fix(<area>): ...` — bug fix
- `refactor(<area>): ...` — non-behavioural change
- `docs(<area>): ...` — markdown only
- `test(<area>): ...` — test-only
- `chore(<area>): ...` — build / tooling / deps
- `ci(<area>): ...` — workflow changes

Reference the issue with `Closes #N` in the body so GitHub auto-closes
on merge. Squash-merge keeps `main` linear.

## Adding a skill

1. Pick a name following `<domain>.<verb>` lower-snake_case.
2. Implement `domain/skills.Skill` under
   `internal/application/<domain>/<name>_skill.go`.
3. Register it in `internal/bootstrap/skills.go`.
4. If it needs a CLI handler, wire it in
   `internal/adapters/inbound/cli/` and add it to `internal/bootstrap/app.go`.
5. Unit tests live next to the skill; contract tests are automatic.
6. Document the new skill in `docs/skills/catalog.md`.

## Adding an outbound adapter

1. New folder under `internal/adapters/outbound/<area>/<name>/`.
2. Implement the relevant `outbound.*Provider` interfaces.
3. Compile-time assertions in a `contract_test.go`.
4. Wire the constructor in `internal/bootstrap/providers.go`.
5. SDK deps go in `go.mod` direct requires (run `go mod tidy`).
6. Add an ADR if the adapter introduces a new dependency family or
   an architectural pattern.

## Pull requests

PRs target `main`. Each PR:

- Has a short, imperative title (`feat(inventory/aws): list S3 buckets`).
- States what changed and why in the body. Tradeoffs > implementation.
- Lists `Closes #N` (or `Refs #N` for partial work).
- Is squash-mergeable (no merge commits in the branch).
- Goes through CI (test + race + lint).
- Gets at least one approval from a maintainer.

## Code review philosophy

- **Smaller is better.** Two focused PRs beat one giant one.
- **Boring code is better than clever code.** Multix is infra; it's
  meant to age well.
- **Tests describe intent.** A change without a test that would have
  caught the bug is incomplete.
- **No premature abstractions.** Three similar lines is fine; a
  half-finished interface is not.

## Reporting issues

Use one of the issue templates:

- Bug — include cloud, multix version, runtime/CLI mode, request ID
  if available.
- Feature — describe the *operator* problem before proposing API.
- Contract regression — label `area/contracts`, P0 turnaround.

## Code of Conduct

By participating you agree to abide by `CODE_OF_CONDUCT.md`.

## License

By contributing you agree your work is released under the same license
as the project.
