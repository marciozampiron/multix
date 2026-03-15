# Multix Skill Rules

## 1. Definition

A skill is a reusable capability of the platform with a clear input/output contract.

A skill must be callable by:
- CLI
- future API
- future AI agents

The skill is the execution unit.
The agent is the orchestration unit.

### Golden Rule
- Agent thinks
- Skill executes

---

## 2. Skill Types

## 2.1 Platform Skill
A platform capability visible at the product level.

Examples:
- doctor
- auth
- inventory
- ai
- k8s
- plugin

## 2.2 Agent Skill
A machine-callable tool-like contract.

Examples:
- doctor.run
- auth.validate
- auth.whoami
- inventory.scan
- inventory.summary
- k8s.list_clusters
- ai.explain
- ai.generate_terraform

---

## 3. Naming Convention

Use dot notation:

`<domain>.<action>`

Examples:
- `doctor.run`
- `auth.validate`
- `inventory.scan`
- `inventory.summary`
- `k8s.list_clusters`
- `ai.explain`
- `ai.generate_terraform`

Rules:
- use lowercase
- use verbs for actions
- prefer stable names
- avoid provider names in the skill name unless absolutely necessary

Bad:
- `aws.inventory.scan`
- `gcp.auth.whoami`
- `myCoolSkill`
- `inventoryDoStuff`

---

## 4. Mandatory Skill Contract

Every skill must implement the project skill interface:

- `Name() string`
- `Description() string`
- `Execute(ctx context.Context, input map[string]any) (map[string]any, error)`

For now, `map[string]any` is acceptable in MVP.
In future versions, evolve to typed input/output models or schemas where useful.

---

## 5. Skill Design Checklist

When creating a new skill:

1. Identify the domain/capability
2. Create or reuse the application use case
3. Keep business logic in `application/` (or `domain/` when justified)
4. Keep CLI handlers thin
5. Register the skill in bootstrap
6. Use machine-friendly output keys
7. Add at least one test
8. Keep it reusable by future agent adapters
9. Avoid provider lock-in when possible

---

## 6. Input/Output Rules

### Inputs
- Keep input keys simple and explicit
- Avoid nested structures unless necessary
- Validate required inputs
- Use stable names

Examples:
- `provider`
- `region`
- `input`
- `resource_type`

### Outputs
- Return machine-friendly maps
- Use stable keys
- Avoid presentation formatting inside the skill

Good:
- `provider`
- `valid`
- `resources`
- `clusters`
- `content`

Bad:
- preformatted table strings
- human-only decorated output
- provider-specific raw SDK dumps

---

## 7. CLI Rule

Cobra command handlers must:
- parse flags/args
- call skill executor
- render output

Cobra command handlers must NOT:
- contain core business logic
- contain provider-specific SDK logic
- duplicate use case logic

---

## 8. Provider Rule

Skills should prefer provider selection through:
- input
- config
- registry lookup

Do NOT hardcode provider selection inside the skill unless explicitly temporary and clearly marked.

---

## 9. Skill Evolution Rule

A skill may evolve from:
- MVP map-based contract
to:
- typed input/output
- JSON schema
- tool manifest
- MCP exposure

But the public semantic behavior should remain stable.

---

## 10. Forbidden

Do NOT:
- put business logic in Cobra commands
- mix prompt engineering with business logic
- make skills depend on UI formatting
- create one-off skills that cannot be reused
- create provider-specific skills unless there is no viable abstraction