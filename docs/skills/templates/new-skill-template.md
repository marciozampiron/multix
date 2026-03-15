# New Skill Template

Use this checklist and template when adding a new skill.

---

## Checklist

- [ ] Skill name follows `<domain>.<action>`
- [ ] Key files have file headers
- [ ] Exported symbols have Go doc comments
- [ ] Business logic lives in `internal/application/`
- [ ] CLI handlers are thin
- [ ] Skill registered in bootstrap
- [ ] At least one test added
- [ ] `docs/skills/catalog.md` updated
- [ ] README example added if user-facing

---

## Suggested Files

- `internal/application/<domain>/<action>.go`
- `internal/adapters/inbound/cli/<domain>.go` (or existing file)
- `internal/bootstrap/skills.go`
- `test/<domain>_<action>_test.go`

---

## Example Skill Contract

```go
// File: internal/domain/skills/example_skill.go
// Company: Hassan
// Creator: Zamp
// Created: DD/MM/YYYY
// Updated: DD/MM/YYYY
// Purpose: Defines the example skill contract for reusable execution flows.

package skills

import "context"

// ExampleSkill executes a reusable capability exposed to CLI and future agents.
type ExampleSkill struct{}

// Name returns the stable machine-readable name of the skill.
func (s *ExampleSkill) Name() string {
	return "example.run"
}

// Description returns the human-readable description of the skill.
func (s *ExampleSkill) Description() string {
	return "Executes the example capability."
}

// Execute runs the skill with the provided input.
func (s *ExampleSkill) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
	return map[string]any{
		"ok": true,
	}, nil
}