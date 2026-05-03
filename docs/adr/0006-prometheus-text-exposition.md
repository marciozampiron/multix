# ADR 0006 — Prometheus Metrics via Text Exposition (no `client_golang`)

## Status
Accepted

## Context
The runtime needs to expose operational metrics — at minimum, skill
invocation counts split by status — so an operator can hook the binary
into the same Prometheus stack that scrapes the rest of their fleet.

The obvious choice is `github.com/prometheus/client_golang`. It pulls
in 30+ transitive dependencies (procfs, expfmt, common, …) that
nothing else in Multix needs. For a CLI binary that brags about a
slim dependency tree, that is a heavy ask for a counter and a handler.

## Decision
Hand-roll the metrics surface against the
[Prometheus text exposition format v0.0.4](https://prometheus.io/docs/instrumenting/exposition_formats/#text-format-version-0-0-4):

- A `Metrics` struct with `sync.Mutex` and a
  `map[metricKey]int` of skill-invocation counts.
- A `Prometheus()` method that walks the map sorted by key and renders
  `# HELP` + `# TYPE` headers and `metric{labels} value` lines.
- A `GET /metrics` handler on the runtime mux that writes the rendered
  text with `Content-Type: text/plain; version=0.0.4`.
- Counter increment happens in the `executeHandler` based on success vs
  error.

## Consequences
- Zero new dependencies. Build size unchanged.
- Histograms and summaries are NOT supported in this implementation —
  if/when latency histograms become a requirement, switch to
  `client_golang`. That migration is local to one file and one ADR
  superseder.
- The exposition format is stable; future Prometheus releases remain
  compatible with v0.0.4.
- Only counters are exposed; gauges (active connections, registry size)
  can be added with the same hand-rolled approach when needed.
