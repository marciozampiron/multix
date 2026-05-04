# Quickstart: Using the MULTIX Agent Runtime

This is the practical quickstart guide for local developers, shell users, AI agent builders, and LLM tool orchestrators.

## Prerequisites

- Go 1.25+ or the version declared in `go.mod`.
- A buildable repository structure.
- Supported providers (`aws`, `gcp`, `oci`) depend on locally configured credentials on the host machine. (e.g., OCI examples assume a valid `~/.oci/config`).

## Build the binary

You can compile the application binary natively:
```bash
go build -o build/multix ./cmd/multix
```
Alternatively, using the canonical Makefile target:
```bash
make build
```

## Start the runtime

To boot the agent execution node:
```bash
./build/multix serve --port 8080
```
- If the `--port` flag is omitted, it defaults to `8080`.
- Logs indicate runtime startup and binding.
- The process runs persistently and blocks until explicitly interrupted (`Ctrl+C`).

## Validate the runtime

### Health
Verify the liveness probe ensuring the server is running.
```bash
curl -s http://localhost:8080/health | jq .
```
*(Without `jq`: `curl -s http://localhost:8080/health`)*

### Tools
Fetch the dynamic manifest list of all skills loaded in the registry.
```bash
curl -s http://localhost:8080/tools
```

### Capabilities
Discover the generic platform capabilities and connected providers.
```bash
curl -s http://localhost:8080/capabilities
```

### Metrics
Scrape Prometheus text exposition metrics.
```bash
curl -s http://localhost:8080/metrics
```

## Execute a skill manually

### Example A - AWS auth validate
Validates if your AWS credentials work end-to-end.
```bash
curl -s -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: quickstart-auth-aws" \
  -d '{
    "skill": "auth.validate",
    "provider": "aws",
    "params": {}
  }'
```

### Example B - OCI auth whoami
Retrieves identifying user boundaries from the OCI profile.
```bash
curl -s -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: quickstart-whoami-oci" \
  -d '{
    "skill": "auth.whoami",
    "provider": "oci"
  }'
```

### Example C - cost.focus_report for unconfigured GCP
This shows the supported-but-unconfigured envelope when Cloud Billing BigQuery export is not configured.
```bash
curl -s -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: quickstart-cost-gcp" \
  -d '{
    "skill": "cost.focus_report",
    "provider": "gcp",
    "params": {
      "providers": ["gcp"],
      "period": "current_month"
    }
  }'
```

### Example D - unknown skill
Displays the explicit 404 response payload when targeting a missing or hypothetical capability.
```bash
curl -i -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "skill": "non.existent.skill"
  }'
```
Expected behavior: the HTTP response maps to `404 Not Found` with an error envelope.

## LLM / agent integration pattern

For AI orchestration clients, we recommend the following discovery and execution loop:

1. **Ping**: Call `GET /health` to ensure the local runtime node is alive.
2. **Discover**: Call `GET /capabilities` to survey the provider support surface exposed by this runtime build.
3. **Parse Specs**: Call `GET /tools` to generate tool-array specs matching standard LLM capability definitions (e.g., matching OpenAI or Anthropic tool spec shapes).
4. **Plan**: Feed the LLM the tools.
5. **Execute**: The LLM issues a deterministic `POST /execute` intent targeting a designated skill.
6. **Evaluate**: Inspect the JSON envelope structure. `ok: true` denotes success.
7. **Reflect**: Retry the tool choice or trigger an alternative if failures arise (e.g., checking other profiles).

## Error handling guidance

Clients should strictly rely on HTTP status codes paired with the `ok` / `error` envelope flags.
- **`200`** = Execution success.
- **`400`** = Malformed JSON payload or missing `skill` string.
- **`404`** = Unknown skill or unavailable execution target.
- **`500`** = Server-side runtime execution failure (e.g., bad auth, network timeout).

**Architectural notes:**
- Clients should **never** parse standard output logs.
- The `provider` key should be sent explicitly as `"provider": "oci"` whenever ambiguity across clouds exists.

## Smoke validation

MULTIX allows strict local CI/CD simulation tests:
```bash
make smoke-oci
```
- Validates the local OCI auth flows end-to-end.
- Confirms graceful failure arrays if `~/.oci/config` is absent from the host.
- Vetoes the pipeline if anomalous stack-trace panics (`^panic:`) occur in the execution layer.
- Designed as a pure deployment readiness guard.

## Troubleshooting

- **Port already in use:** Verify nothing else on the host listens to `8080`.
- **OCI config missing or invalid:** The execution results in `ok: false` returning `bad configuration` or `failed to retrieve` in the JSON response payload payload.
- **Unknown skill returns 404:** Check for typos in the `"skill": ""` parameter matching the exact dot-notation schema.
- **Malformed JSON returns 400:** Make sure trailing commas or unquoted strings aren't polluting the JSON POST blob.
- **Empty /tools result:** Usually denotes zero registered capabilities wired inside `internal/bootstrap/skills.go`.
- **GCP billing not configured:** Set `MULTIX_GCP_BILLING_DATASET=<project>.<dataset>.<table>`.
- **OCI billing not configured:** Set `MULTIX_OCI_TENANCY_OCID=<tenancy_ocid>`.

## Stability note

The current local runtime endpoints are the practical reference surface for agent and tool-calling clients. The future MCP/plugin model is documented in ADR 0008 and remains implementation work under issue #20.
