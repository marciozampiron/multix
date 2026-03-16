# MULTIX Runtime Architecture (v0.4-alpha)

This is the authoritative architecture document for the local HTTP runtime adapter introduced in v0.4. The runtime exposes an agent-facing execution surface while keeping the CLI entrypoint extremely thin. It is an intentional adapter-based design kept maximally minimal for the v0.4 milestone.

## Why this exists

The `multix serve` runtime was introduced to:
- Expose skills over HTTP natively for AI agents, LLM tool orchestrators, and external automation.
- Preserve the existing CLI-first design while enabling machine-to-machine execution.
- Avoid overengineering by keeping the `ToolAdapter` as the stable bridge across all surfaces.
- Strictly maintain a local-first operational model before engaging any remote or distributed daemon deployment concepts.

## High-level architecture

The program boots from `cmd/multix/main.go`, performing bootstrap wiring and Cobra command registration. The `serve` command triggers the runtime adapter (`internal/adapters/inbound/runtime/server.go`), opening an `http.ServeMux` to handle inbound API traffic. Handlers decode JSON envelopes and delegate strictly to the unified `ToolAdapter`, which maps against the internal skill registry and executor, eventually dispatching to provider-aware domain logic.

```text
CLI (cobra)
  -> serve command
    -> runtime.Server
      -> http.ServeMux
        -> /health
        -> /tools
        -> /execute
        -> /capabilities
          -> ToolAdapter
            -> skill registry / executor
              -> provider-aware skill implementations
```

## Design principles

- **Standard library first**: Strictly leveraging Go's native `net/http` and v1.22+ routing patterns.
- **No web framework dependency**: Avoids Gin, Echo, Fiber, or massive framework footprints.
- **Thin inbound CLI adapter**: The generic Cobra entrypoint has almost no logic beyond flag binding.
- **Dedicated inbound runtime adapter**: HTTP encoding/decoding is isolated and strictly typed.
- **Reuse over reinvention**: Relies exactly on the same `ToolAdapter` the CLI and LLM Manifests use.
- **Explicit HTTP contracts**: Strict request boundaries and response JSON blocks.
- **Graceful shutdown**: Correct signal interception prevents orphaned processes.
- **JSON-first machine interface**: Clean serialization optimized for machine intelligibility.
- **Local-first execution model**: Meant for local sandboxing and safe terminal contexts.
- **Intentional Simplicity**: v0.4 intentionally avoids auth middleware, persistence, distributed runtime concerns, or plugin overengineering.

## Current runtime endpoints

### GET /health
Returns the basic liveness probe of the runtime engine.
- **Method:** `GET`
- **Path:** `/health`
- **Purpose:** Verifies that the server process is alive and responding.
- **Response Shape:**
```json
{
  "status": "ok",
  "service": "multix",
  "mode": "runtime"
}
```
- **Status Codes:** `200 OK`

### GET /tools
Returns tool manifests generated dynamically directly from the internal skill registry.
- **Method:** `GET`
- **Path:** `/tools`
- **Purpose:** Exposes a machine-discoverable JSON contract mapping the capabilities that AI Agents can pipe into function-calling setups. Reflects the *active registry*, not a hardcoded static catalogue.
- **Response Shape:** array of tool descriptors (name, description, parameters).
- **Status Codes:** `200 OK`

### POST /execute
The master orchestration endpoint. Expects a JSON envelope dictating which skill to run and its parameters.
- **Method:** `POST`
- **Path:** `/execute`
- **Purpose:** Triggers domain skills securely over HTTP.
- **Request Body:**
```json
{
  "skill": "auth.validate",
  "provider": "oci",
  "params": {
    "profile": "DEFAULT"
  }
}
```

**Execution Semantics:**
- The `skill` key is strictly required. 
- The `provider` key is optional but, when present, is globally injected into `params["provider"]` if not already present.
- Underlying execution strictly delegates to `ToolAdapter.Execute(ctx, skill, params)`.

**Response (Success Envelope):** `200 OK`
*Note: The `result` payload is skill-specific and its shape varies by skill. The response body below is illustrative, not a fixed canonical schema.*
```json
{
  "ok": true,
  "skill": "auth.validate",
  "provider": "oci",
  "result": { 
    "caller_identity": "ocid1...",
    "is_valid": true 
  }
}
```

**Response (Error Envelope - Unknown Skill):** `404 Not Found`
When the underlying execution path reports an explicit "not found" error.
```json
{
  "ok": false,
  "error": "failed to execute skill 'foo': skill not found"
}
```

**Response (Error Envelope - Failures):** `400 Bad Request` | `500 Internal Server Error`
Malformed JSON or missing the `skill` string yields `400`. Runtime provider errors (e.g., bad auth headers from remote cloud) yield `500`.
```json
{
  "ok": false,
  "error": "failed to validate OCI credentials: authorization failed"
}
```

### GET /capabilities
Returns the runtime capability matrix mapping enabled features and declared runtime-supported providers.
- **Method:** `GET`
- **Path:** `/capabilities`
- **Purpose:** A lightweight discovery endpoint for orchestration clients.
- **Response Shape:**
```json
{
  "api_version": "v1",
  "capabilities": [
    "tool_execution",
    "dynamic_manifests"
  ],
  "supported_providers": ["aws", "gcp", "oci"]
}
```
- **Status Codes:** `200 OK`

## Runtime lifecycle

To boot the engine: `multix serve --port 8080`.
The server starts logging and binds to the specified port. 

A critical element of the architecture is its *Graceful Shutdown* handling. Upon receiving `os.Interrupt` or `SIGTERM`, the runtime enforces a strict 5-second shutdown timeout context before forcefully terminating. This ensures that in-flight skill executions wrap up safely, which is imperative for CI workflows, containerization limits (like Docker's `STOPSIGNAL`), and future service wrappers.

## Testing strategy (v0.4)

The current testing regime rigorously dictates:
- **Unit tests for runtime handlers:** Verifying parsing and serialization of `executeRequest` and envelopes inside `server_test.go`.
- **Endpoint coverage:** Mocks and tests assert coverage for `/health`, `/tools`, and `/execute`.
- **404 Mapping & Provider Injection:** The exact behavior of routing unknown skills to `404` and enforcing `params["provider"]` logic is unit-tested sequentially.
- **OCI Smoke Script for auth flows:** End-to-end execution of the binary locally via `scripts/e2e/oci_smoke.sh` hooked up by `make smoke-oci`. These smoke tests are *intentionally strict*, validating graceful failure when `~/.oci/config` is absent and demanding absolute lack of `panic` logs from Go core. They are specifically not full-blown cloud integration suites, but bulletproof deployment readiness guards.

## Known constraints / intentional non-goals

The following are strictly out of scope for the current design:
- No auth middleware or token checks on runtime endpoints.
- No TLS termination in-process (must be handled by a proxy if exposed).
- No OpenAPI (Swagger) spec generation for the runtime yet.
- No request tracing / telemetry middleware yet.
- No Prometheus metrics endpoint yet (planned later).
- No remote daemon deployment contract yet (assumes `localhost` binding).
- No streaming execution protocol yet (WebSocket/SSE).
- No plugin sandboxing yet (WASM/shared object isolation).

## Forward path

The runtime paves the exact groundwork required for broader integrations:
- **v0.5**: Broad inventory expansion parsing EC2/S3, GCP Compute, and OCI Object Storage payloads back through the Execute path.
- **v0.6**: Observability enhancements tracking capability invocations.
- **v0.7**: Higher-level doctor skills (`doctor.auth`).
- **v1.0**: Absolute contract stabilization and public plugin extension models.
