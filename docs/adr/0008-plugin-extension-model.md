# ADR 0008 — Plugin Extension Model

## Status
Proposed

## Context
The community needs a way to ship skills outside the main repository
without forking. The naive answer is "write a Go plugin" — but Go's
native `plugin` package has well-known and documented limits:

- Linux + macOS only (no Windows, no static cross-compile).
- Plugin and host MUST be built with byte-for-byte identical Go
  toolchains, modules and build tags. A patch-level Go upgrade on
  the host invalidates every shipped plugin.
- No unload, no reload, no per-tenant isolation.
- The `plugin` package is in long-term maintenance mode upstream —
  the Go team explicitly recommends alternatives.

Three real alternatives were considered:

### Option A — In-tree skill SDK + tagged builds
Skills live in their own Go modules; the Multix binary is rebuilt
with a tag per consumer. No protocol, easiest to implement.

- ✅ Compile-time safety, zero runtime overhead.
- ❌ Requires every plugin user to maintain their own binary.
- ❌ Doesn't solve "third-party skills" — only "first-party
  selection of which skills to bundle".

### Option B — Subprocess + RPC (e.g. `hashicorp/go-plugin`)
Plugins run as separate processes; Multix communicates via gRPC over
a Unix socket. Same pattern as Terraform providers and Vault plugins.

- ✅ Process isolation, hot crash recovery.
- ✅ Plugin author can use any language that targets gRPC.
- ❌ ~30 transitive dependencies just for the plugin runtime.
- ❌ Operational complexity (lifecycle, logs, port allocation,
  permissions on the socket).

### Option C — Federated HTTP/MCP servers
Plugins are independent HTTP servers that speak the
[Model Context Protocol (MCP)](https://modelcontextprotocol.io)
or the canonical Multix tool-manifest shape (ADR 0007). The Multix
runtime knows how to discover and proxy them.

- ✅ Multix already speaks HTTP+JSON. The proxy is a thin layer.
- ✅ Plugins can be in any language — Python, TypeScript, Rust.
- ✅ MCP is rapidly becoming the de-facto LLM tool standard;
  reusing it means every Multix plugin is also usable from any
  MCP-compatible client (Claude Desktop, Cursor, etc.).
- ✅ Operationally identical to running Multix itself
  (`multix serve` ↔ `<plugin> serve`).
- ❌ Network latency per call (acceptable: skills are I/O-bound on
  cloud SDKs anyway).
- ❌ Requires a discovery/registration story (covered below).

## Decision
Adopt **Option C: Federated HTTP/MCP servers**.

Multix gains two new responsibilities:

1. **Be an MCP server**: existing skill registry exposed via the
   MCP wire protocol on `multix serve --mcp`. Any MCP client
   (Claude Desktop, OpenAI Tools, Gemini Function Calling adapters)
   can list and invoke Multix skills with zero plugin work.

2. **Be an MCP client**: a new `[plugins]` block in the runtime
   config lists upstream MCP servers. At bootstrap, Multix queries
   each one's manifest, prefixes their tool names with the plugin's
   logical name (e.g. `mycorp/inventory.scan_iac`) and registers
   them as proxy skills in the local Skill registry.

   ```toml
   [plugins.mycorp]
   url = "http://localhost:9000"
   protocol = "mcp"           # or "multix" for native canonical shape
   timeout_seconds = 30
   ```

Local skills retain their namespace; remote ones get their plugin
name as a prefix to prevent collisions.

Native Go `plugin` is **rejected** — none of its tradeoffs are worth
the lock-in.

`hashicorp/go-plugin` is **deferred** — if a future need for
in-process performance and process isolation appears, it can be
added as Option D without superseding C.

## Consequences

### Positive
- Plugin authors write a normal HTTP server in any language — no
  Go required, no ABI lock-in.
- Multix integrates with the broader MCP ecosystem (Claude Desktop,
  Cursor, custom agents) the moment it implements MCP server.
- The `Skill` interface (#19) remains the single contract;
  proxy-skills implement it by forwarding to the upstream server.
- Operational story is consistent: every plugin is a process
  exposing the same surface as the host.

### Negative / costs
- Need to ship the MCP server adapter (~1 week of work).
- Need to ship the MCP client / plugin loader (~1 week).
- Plugin discovery requires an explicit config block — there is no
  auto-discovery in v1, by design (security).
- Network failures must be surfaced clearly; a dead plugin should
  not block local skills.

### Migration / sequencing
1. Implement MCP server first (`multix serve --mcp`). This unlocks
   value for users without any plugins of their own.
2. Implement plugin loader as a follow-up issue.
3. Document a "build your first plugin" walkthrough using TypeScript
   + the official MCP SDK.

### Open questions
- Authentication between Multix and a remote plugin (mutual TLS?
  shared secret?). v1 assumes loopback only; remote plugins are
  v1.5+.
- Schema validation: should Multix re-validate every plugin's
  `inputSchema` against the contract test? (Recommendation: yes —
  the contract is universal.)
- Versioning: how does Multix react when a plugin's manifest
  changes between launches? (Recommendation: warn but tolerate;
  fail only if an in-flight skill name disappears.)

## References
- ADR 0001 — Skills-first architecture (the proxy-skill pattern
  reuses the same contract).
- ADR 0007 — Tool manifest shape (the canonical descriptor that the
  MCP server adapter renders).
- [Model Context Protocol spec](https://modelcontextprotocol.io)
- [Go `plugin` package limitations](https://pkg.go.dev/plugin#hdr-Limitations)
