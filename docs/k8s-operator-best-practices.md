# Kubernetes Operator Best Practices for Multix

This guide is for **operators** running Multix against managed Kubernetes
clusters (EKS, GKE, OKE) — either interactively from a workstation or
inside a pipeline that drives the runtime over HTTP.

It is opinionated: every recommendation here has a "why" anchored in the
runtime's architecture (`docs/runtime-architecture.md`) and the v1.0
contracts (`docs/standards/contracts.md`).

## 1. Authentication model: assume, don't embed

Multix never stores credentials. Every cloud SDK call uses the ambient
credential chain of the host process. For cluster-touching skills:

- **AWS/EKS**: prefer SSO (`aws sso login --profile <prod>`) or
  IRSA when running inside a pod. Avoid long-lived IAM user keys —
  `security.identity_posture` will flag them as `medium`.
- **GCP/GKE**: prefer Workload Identity (in-cluster) or
  `gcloud auth application-default login` (workstation). A
  service-account JSON key on disk is a P1 finding.
- **OCI/OKE**: prefer Instance Principal when running on OCI
  Compute; fall back to `~/.oci/config` with a session token,
  not a long-lived API key.

Run `doctor.auth` before any session-modifying skill — it returns one
finding per provider. Consume the `findings[].ok` field; abort the run
if any provider you intend to touch is `false`.

## 2. Cluster discovery: list before you act

`k8s.list_clusters` is **read-only and free** for all three providers.
Always call it first to verify:

- The expected number of clusters returned.
- Lifecycle state (`ACTIVE` for AWS/OCI, `RUNNING` for GCP) — anything
  else means the cluster is undergoing a control-plane operation and
  should not be the target of subsequent skills.
- Version drift across the fleet — if `version` differs by more than
  one minor across clusters, document the upgrade gap.

`doctor.k8s` rolls these checks into a single call and returns a
`total_healthy` aggregate. CI gates should fail when
`total_healthy < total_clusters`.

## 3. Context sync: never assume `kubectl` follows Multix

`SyncContext(clusterName, region)` is a **side effect on the local
kubeconfig**. It is not invoked automatically. Operators must:

- Call it explicitly when they want kubectl to follow the cluster
  Multix just listed.
- Treat the kubeconfig as user state — Multix never overwrites an
  existing context without explicit invocation.
- Audit `~/.kube/config` after pipeline runs; CI machines should
  point `KUBECONFIG` to a scratch path so syncs don't leak across
  jobs.

## 4. Multi-cluster orchestration: one provider per request

Skills that take a `provider` parameter are scoped to one cloud. To act
across clouds, call the skill three times in sequence (one per
provider) and aggregate client-side. The runtime intentionally does
NOT have a "multi-provider" execution mode — that orchestration belongs
in the agent loop or the pipeline.

For audit-style skills that accept `providers[]` (`doctor.auth`,
`doctor.k8s`, `landingzone.audit`, `security.identity_posture`,
`cost.quick_scan`), the runtime fans out internally and returns a
single response. Use that variant when you need a coherent snapshot
across clouds.

## 5. Region and compartment scoping

| Provider | What gets scanned                                                |
|----------|-------------------------------------------------------------------|
| AWS      | Active region from SDK config (`AWS_REGION` / profile)           |
| GCP      | All locations for the project (locations="-")                   |
| OCI      | Tenancy compartment (root); sub-compartments require explicit ID |

If you run multi-region on AWS, run the EC2 listing once per region —
the adapter does not iterate regions automatically. A future
enhancement (`#region-aware-iteration`) is tracked outside this guide.

## 6. Logging, request IDs and correlation

Every runtime request gets an `X-Request-ID` (ADR 0005). For pipeline
use:

- Set the header explicitly to your CI job ID so logs link back to
  the run.
- Read it back from the response and stamp it on any artifacts
  (Slack messages, ticket comments).
- The same ID is included in `params["request_id"]` and in every
  structured log line under `request_id=...`.

## 7. Metrics: scrape locally, don't expose

`GET /metrics` (ADR 0006) is loopback-only by default. Do not expose
the runtime port to a public network without a reverse proxy that
strips `/metrics` from external traffic. Recommended scrape config:

```yaml
- job_name: multix-local
  static_configs:
    - targets: ["127.0.0.1:8080"]
  scrape_interval: 30s
```

A failed skill increments `multix_skill_invocations_total{status="error"}`
by one. Alert on `rate(...)` rather than absolute count.

## 8. Failure modes and remediation

| Symptom                                              | Likely cause                                                    | Fix                                                             |
|------------------------------------------------------|------------------------------------------------------------------|-----------------------------------------------------------------|
| `auth.validate` returns `valid=false` for AWS        | Expired SSO session                                             | `aws sso login --profile <prod>`                               |
| `k8s.list_clusters` returns empty for GCP            | ADC has no project or cluster API disabled                      | `gcloud config set project <id>` + enable Container API        |
| OCI list returns "missing tenancy"                   | `~/.oci/config` invalid                                         | Reconfigure via `oci session authenticate`                     |
| Skill execution hangs                                | Provider SDK retry storm (most common: AWS throttling)          | Inspect logs by request_id; reduce parallelism in agent loop   |

## 9. Upgrade path

Skills are versioned by name (ADR 0007 / contract test). When a skill's
behaviour changes incompatibly, a new skill name is introduced (e.g.
`inventory.scan` → `inventory.scan_v2`). Operators control the cutover
by switching the name they call — Multix never silently changes
semantics behind a stable name.

## 10. What this guide does NOT cover

- Operator pattern for **deploying Multix as a Kubernetes Operator
  CRD** (Multix orchestrating CRDs vs. running inside a cluster). That
  is tracked under `#multix-as-operator` and is outside v1.0 scope.
- Production-grade auth on the HTTP runtime — see ADR 0004 for the
  reverse-proxy guidance until v1.5 ships native auth.
