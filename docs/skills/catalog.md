# Skills Catalog

The active skill registry is built in `internal/bootstrap/skills.go`. Skill names are stable v1.0 contract identifiers and must keep the `<domain>.<action>` format.

## Registered Skills

### AI

- `ai.explain`

### Auth

- `auth.login`
- `auth.validate`
- `auth.whoami`

### Cost

- `cost.focus_report`
- `cost.quick_scan`

### Doctor

- `doctor.auth`
- `doctor.k8s`
- `doctor.run`

### Infrastructure

- `infra.generate_network`

### Inventory

- `inventory.scan`
- `inventory.summary`

### Kubernetes

- `k8s.list_clusters`

### Landing Zone

- `landingzone.audit`

### Security

- `security.k8s_audit`
- `security.iam_audit`
- `security.identity_posture`

---

## Provider Notes

- AWS: auth, inventory, Kubernetes, security, and cost paths use AWS SDK adapters where implemented.
- GCP: auth, inventory, Kubernetes, and cost paths use Google SDK adapters where implemented.
- OCI: auth, inventory, Kubernetes, and cost paths use OCI SDK adapters where implemented.
- AI: Gemini is the current AI provider path.

`cost.focus_report` uses:

- AWS Cost Explorer.
- GCP Cloud Billing BigQuery export through `MULTIX_GCP_BILLING_DATASET=<project>.<dataset>.<table>`.
- OCI Usage API through `MULTIX_OCI_TENANCY_OCID=<tenancy_ocid>`.

## Rules

- Keep names stable once shipped.
- Use `<domain>.<action>`.
- Prefer provider-agnostic names.
- Register each skill exactly once in `internal/bootstrap/skills.go`.
- Document new skills here when added.
- Keep business behavior inside application skills, not Cobra handlers.
