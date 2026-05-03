# Multix Documentation

This directory is the documentation hub for the Multix skills-first multi-cloud runtime.

## Start Here

- `../README.md`: product overview, quick start, runtime endpoints, provider notes.
- `quickstart-agent-runtime.md`: hands-on guide for `multix serve` and `/execute`.
- `testing-playbook.md`: caderno de teste for local validation, runtime smoke tests, provider checks, and release gates.
- `skills/catalog.md`: current registered skills and provider notes.
- `runtime-architecture.md`: HTTP runtime architecture and endpoint contracts.

## Architecture

- `adr/0001-skills-first-architecture.md`
- `adr/0002-provider-registry.md`
- `adr/0003-agent-tool-contracts.md`
- `adr/0004-local-http-runtime.md`
- `adr/0005-request-id-propagation.md`
- `adr/0006-prometheus-text-exposition.md`
- `adr/0007-tool-manifest-shape.md`
- `adr/0008-plugin-extension-model.md`

ADR 0008 is the source of truth for the proposed plugin extension model. It is design-complete, but implementation remains tracked under issue #20.

## Standards

- `standards/contracts.md`: v1.0 contract promise for skills and ports.
- `standards/coding-style.md`: Go coding conventions.
- `standards/doc-comments.md`: comment expectations.
- `standards/file-headers.md`: file header conventions.

## Templates

- `templates/go-file-template.md`
- `skills/templates/new-skill-template.md`

## Project Notes

- `project-historical-resolution.md`: audit trail for Project items resolved before a PR was linked.
- `../project/PROJECT_ROADMAP.md`: roadmap status and operating notes.

Read `../CONTRIBUTING.md`, `standards/contracts.md`, and `testing-playbook.md` before opening a pull request.
