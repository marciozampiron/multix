# Project Historical Resolution Register

This register records Project items that were closed before a GitHub pull request
was attached to the issue. GitHub Projects populates the `linked pull requests`
field automatically from real PR relationships; it cannot be backfilled with a
commit-only closure.

Use this file when auditing the `MULTIX Product Roadmap` project and deciding
whether an older item was resolved by a direct commit, a squash merge outside a
PR, or a duplicate/consolidation event.

## v0.4 Historical Items

| Issue | Resolution source | Notes |
| --- | --- | --- |
| #3 | Manual duplicate closure | Closed on 2026-03-16 with `Closed as duplicate due to script retry.` No commit or PR was linked in the timeline. |
| #4 | Commit `4a69f0e` | `feat(oci): implement OCI Auth Provider with identity validation` closed the issue on 2026-03-16. |
| #5 | Commit `d75092c` | `feat(runtime): add multix serve local runtime base and /health endpoint` closed the issue on 2026-03-16. |
| #6 | Commit `d75092c` | Same runtime base and health endpoint commit closed the issue on 2026-03-16. |
| #7 | Commit `ef83794` | `feat(runtime): expose POST /execute and GET /capabilities` closed the issue on 2026-03-16. Follow-up commits `5c3eba7` and `b4af7e1` also referenced it. |
| #25 | Squash merge note for `docs-runtime-quickstart` | Timeline referenced commits `4564a90`, `63a1ad5`, and `be0c9d9`, then closed with the note `Closed via squash merge of docs-runtime-quickstart`. |
| #26 | Squash merge note for `docs-runtime-quickstart` | Same documentation flow as #25: commits `4564a90`, `63a1ad5`, and `be0c9d9`, followed by the squash merge closure note. |
| #34 | Manual duplicate closure | Closed on 2026-03-16 as duplicate of #27 to consolidate v0.4 CI work. |
| #40 | Commit `4a69f0e` | `feat(oci): implement OCI Auth Provider with identity validation` closed the issue on 2026-03-16. |
| #41 | Commit `671f120` | `feat(runtime): expose /tools endpoint for dynamic agent manifests` closed the issue on 2026-03-16. |
| #42 | Commit `ef83794` | `feat(runtime): expose POST /execute and GET /capabilities` closed the issue on 2026-03-16. |

## Project Handling

For historical closures without PRs, add an issue comment with the resolver
source before moving the Project item to `Done`. For new work, close issues
through a normal PR so GitHub can attach the pull request automatically.
