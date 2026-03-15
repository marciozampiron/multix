# Multix Testing Rules

## 1. Testing Philosophy

Tests should verify behavior, not implementation trivia.

Goals:
- confidence
- regression prevention
- stable evolution
- lightweight maintenance

---

## 2. MVP Testing Priorities

Priority order:
1. skill registry
2. skill executor
3. core use cases
4. provider stubs
5. CLI smoke paths (selectively)

---

## 3. Preferred Test Style

Use:
- table-driven tests when useful
- focused unit tests
- smoke tests for core wiring
- minimal mocks

Avoid:
- over-mocking
- brittle tests tied to internals
- snapshot tests for unstable outputs

---

## 4. What to Test for Each New Skill

For each new skill, prefer at least:

1. registration test
2. successful execution test
3. invalid input test (if applicable)

If the skill has branching:
- use table-driven tests

---

## 5. What Not to Test Excessively

Do NOT over-test:
- trivial getters
- obvious wiring if already covered by smoke tests
- private helper minutiae unless complex

---

## 6. Test Naming

Use descriptive names:
- `TestSkillRegistryRegistersCoreSkills`
- `TestExecuteDoctorSkill`
- `TestInventoryScanReturnsResources`
- `TestAIExplainRequiresInput`

---

## 7. Commands

Preferred:
- `make test`
- `make test-race`

Before finalizing changes:
- run `make fmt`
- run `make vet`
- run `make test`