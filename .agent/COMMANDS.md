# Multix Commands & Workflow Rules

## 1. Preferred Commands

Always prefer the repository Makefile over ad-hoc shell commands.

Primary commands:
- `make build`
- `make run`
- `make test`
- `make test-race`
- `make fmt`
- `make vet`
- `make vuln`
- `make tidy`
- `make lint`

---

## 2. Standard Development Flow

Typical safe workflow:
1. inspect current files
2. make focused changes
3. run:
   - `make fmt`
   - `make vet`
   - `make test`
4. if dependency changes:
   - `make tidy`
5. if security-sensitive:
   - `make vuln`

---

## 3. Forbidden Behavior

Do NOT:
- run destructive shell commands without explicit request
- remove unrelated files
- mass-rename packages without strong reason
- modify Makefile conventions casually
- introduce new build systems without explicit approval