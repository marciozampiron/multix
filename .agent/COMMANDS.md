# Multix Commands & Workflow Rules

## 1. Preferred Commands

Always prefer the repository Makefile over ad-hoc shell commands.

Primary commands:
- `make all`
- `make build`
- `make build-cross`
- `make run`
- `make test`
- `make test-race`
- `make fmt`
- `make vet`
- `make tidy`

---

## 2. Standard Development Flow

Typical safe workflow:
1. inspect current files
2. make focused changes
3. run:
   - `go test ./...`
   - `go vet ./...`
   - `make all`
   - `make build-cross`
   - `git diff --check`
4. if touching runtime, registry, adapters, or shared state:
   - `make test-race`
5. clean generated build artifacts unless they are part of the requested change
6. commit changes on a focused branch
7. open a GitHub pull request before considering the activity delivered

See also:
- `.agent/PULL_REQUESTS.md`
- `docs/testing-playbook.md`

---

## 3. Forbidden Behavior

Do NOT:
- run destructive shell commands without explicit request
- remove unrelated files
- mass-rename packages without strong reason
- modify Makefile conventions casually
- introduce new build systems without explicit approval
