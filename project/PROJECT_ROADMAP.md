# MULTIX Product Roadmap

This roadmap mirrors the GitHub Project `MULTIX Product Roadmap` and the repository state after the v1.0 beta hardening work.

## Product Direction

MULTIX is a skills-first multi-cloud runtime:

- CLI for humans.
- Local HTTP runtime for agents and automation.
- Stable skill and provider contracts.
- Future MCP-based plugin federation.

## Delivered

- v0.1 foundation: Go module, Cobra, Makefile, repository standards.
- v0.2 skills-first runtime: skill contract, registry, executor, agent tool manifests.
- v0.3 real auth: AWS and GCP validation/whoami paths.
- v0.4 runtime: `multix serve`, `/health`, `/tools`, `/execute`, `/capabilities`.
- v0.5 inventory: AWS/GCP/OCI inventory paths.
- v0.6 Kubernetes: EKS/GKE/OKE list-cluster paths.
- v0.7 operational skills: doctor, landing zone, identity posture, cost quick scan.
- v0.8/v0.10 AI-assisted skills: network generation, Kubernetes audit, IAM audit.
- v0.9 FinOps: FOCUS-aligned `cost.focus_report` for AWS, GCP, and OCI.
- v1.0 stabilization: public contracts, ADRs, governance docs, CI hardening.

## Open

- #20 Plugin story and extension model implementation.
- #21 Enterprise Skill Catalog decision remains intentionally deferred unless reopened by the maintainer.

## Governance Rules

- Every non-trivial change ships through a branch and PR.
- PR bodies should include summary, decisions, validation, and linked issue.
- Use `Closes #N` only when the issue is fully delivered.
- Use `Refs #N` or `Part of #N` for design-only or partial work.
- Do not mark Project items `Done` until the closing PR is merged.
- Keep ADRs historical; add a new ADR for a new architecture decision instead of rewriting an accepted one.

## Validation Gate

Minimum local gate before PR:

```bash
go test ./...
go vet ./...
make all
make build-cross
git diff --check
```

Use `make test-race` for shared runtime, registry, adapter, or concurrency-sensitive changes.

## Stretch Vision

- MCP server exposing the local skill registry.
- MCP client/plugin loader for external tool servers.
- Namespaced remote tools with manifest validation.
- Provider-aware golden paths for enterprise operations.
