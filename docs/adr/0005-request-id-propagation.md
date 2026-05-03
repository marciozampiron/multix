# ADR 0005 — Request ID Propagation

## Status
Accepted

## Context
A single agent turn often invokes 3–10 skills in rapid succession. When
something fails, the operator needs to correlate logs, the agent
transcript, and the upstream Cloud SDK call within a single trace
identifier. The old runtime had no notion of a per-request ID, so log
lines from concurrent skills interleaved without affinity.

## Decision
Every inbound HTTP request flows through a `withRequestContext`
middleware in `internal/adapters/inbound/runtime/server.go` that:

- Reads `X-Request-ID` from the request header. Generates a UUIDv4 if
  absent.
- Echoes the value back in the response `X-Request-ID` header.
- Stores it on `context.Context` under an unexported key, exposed via
  `RequestIDFromContext(ctx)`.
- Attaches `request_id` as a structured-log field for "request
  started" / "request completed" lines.
- For `POST /execute`, injects the ID into `params["request_id"]` so
  downstream skill code can propagate it into outgoing SDK calls (e.g.
  via AWS request user-agent strings).

## Consequences
- An external orchestrator that already mints trace IDs can reuse them
  by setting the header — no new ID is generated.
- Adapter authors can pull the ID from `params["request_id"]` and set
  it on their SDK clients so cloud-side audit logs link back to the
  Multix run.
- The middleware is the only writer of the context key, so skill code
  cannot accidentally mutate it.
- A future migration to OpenTelemetry trace IDs is mechanical: replace
  the UUID generator with the OTel `traceparent` parser.
