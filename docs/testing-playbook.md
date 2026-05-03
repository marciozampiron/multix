# MULTIX Testing Playbook

This is the caderno de teste for validating MULTIX locally before opening or merging a PR.

## 1. Baseline

Start from a clean tree:

```bash
git status --short --branch
```

Expected:

```text
## <branch>...origin/<branch>
```

No modified or untracked files should appear unless they are part of the PR.

## 2. Toolchain

Check Go and module metadata:

```bash
go version
go env GOPATH GOMOD
go list ./...
```

Expected:

- Go version is compatible with `go.mod`.
- `GOMOD` points to this repository's `go.mod`.
- `go list ./...` exits successfully.

## 3. Unit And Contract Tests

Run the same core checks used by CI:

```bash
test -z "$(gofmt -l .)"
go test ./...
go vet ./...
```

Expected:

- No gofmt output.
- All packages pass tests.
- Vet exits cleanly.

## 4. Full Local Gate

Run the repository gate:

```bash
make all
make build-cross
git diff --check
```

Expected:

- `make all` runs clean, fmt, vet, tidy, test, and native build.
- `make build-cross` creates Linux and Darwin amd64/arm64 binaries.
- `git diff --check` reports no whitespace errors.

`make all` and `make build-cross` create files under `build/`. If they are not part of your PR, clean them before committing:

```bash
rm build/multix-darwin-amd64 build/multix-darwin-arm64 build/multix-linux-amd64 build/multix-linux-arm64
git checkout -- build/multix
```

## 5. Race Gate

Run this for changes touching runtime, registry, adapters, provider clients, request IDs, or shared state:

```bash
make test-race
```

Expected:

- All tests pass with the race detector.
- No data race report appears.

## 6. Runtime Smoke Test

Build and start the local runtime:

```bash
make build
./build/multix serve --port 8080
```

In another terminal:

```bash
curl -s http://localhost:8080/health
curl -s http://localhost:8080/tools
curl -s http://localhost:8080/capabilities
curl -s http://localhost:8080/metrics
```

Expected:

- `/health` returns service status.
- `/tools` returns registered tool manifests.
- `/capabilities` returns runtime capabilities and providers.
- `/metrics` returns Prometheus text exposition.

## 7. Execute Smoke Test

Run a request through `/execute`:

```bash
curl -s -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: test-auth-validate" \
  -d '{
    "skill": "auth.validate",
    "provider": "aws",
    "params": {}
  }'
```

Expected:

- The response is JSON.
- The runtime does not panic.
- If local cloud credentials are absent, the error is explicit and contained in the response envelope.

Unknown skill check:

```bash
curl -i -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{"skill":"missing.skill"}'
```

Expected:

- HTTP status is `404`.
- Response body is an error envelope.

## 8. Cost Focus Report Checks

AWS:

```bash
./build/multix auth validate --provider aws --output json
curl -s -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "skill": "cost.focus_report",
    "params": {
      "providers": ["aws"],
      "period": "current_month"
    }
  }'
```

GCP without billing export configured:

```bash
unset MULTIX_GCP_BILLING_DATASET
curl -s -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "skill": "cost.focus_report",
    "params": {
      "providers": ["gcp"],
      "period": "current_month"
    }
  }'
```

Expected:

- Provider is `gcp`.
- `supported` is true.
- `reason` explains that BigQuery export is not configured.

GCP with billing export configured:

```bash
export MULTIX_GCP_BILLING_DATASET=<project>.<dataset>.<table>
```

Prerequisites:

- Cloud Billing BigQuery export enabled.
- Runtime identity has `roles/bigquery.dataViewer` on the export dataset.

OCI without tenancy configured:

```bash
unset MULTIX_OCI_TENANCY_OCID
curl -s -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "skill": "cost.focus_report",
    "params": {
      "providers": ["oci"],
      "period": "current_month"
    }
  }'
```

Expected:

- Provider is `oci`.
- `supported` is true.
- `reason` explains that `MULTIX_OCI_TENANCY_OCID` is not configured.

OCI with Usage API configured:

```bash
export MULTIX_OCI_TENANCY_OCID=<tenancy_ocid>
```

Prerequisites:

- Cost and Usage Reports enabled.
- IAM allows usage-budgets read access in tenancy.

## 9. CLI Smoke Matrix

These commands should render JSON or explicit provider errors:

```bash
./build/multix version
./build/multix auth validate --provider aws --output json
./build/multix auth whoami --provider gcp --output json
./build/multix inventory list --provider aws --service compute --output json
./build/multix inventory list --provider gcp --service storage --output json
./build/multix inventory list --provider oci --service compute --output json
./build/multix k8s clusters --provider aws --output json
./build/multix ai explain "Explain this CLI" --output json
```

Expected:

- Commands do not panic.
- Missing local credentials produce actionable errors.
- JSON output stays machine-readable.

## 10. PR Readiness

Before pushing:

```bash
git status --short --branch
git diff --stat
git diff --check
```

PR body should include:

- Summary of changed docs or behavior.
- Decisions and guard-rails.
- Validation commands run.
- `Closes #N` only when the issue is fully delivered.
- `Refs #N` or `Part of #N` for partial or design-only work.
