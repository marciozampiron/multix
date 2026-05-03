# Multix Pull Request Rules

## 1. Default Delivery Rule

Every completed implementation activity must be delivered through a GitHub pull
request.

Do not treat local commits, local branch changes, or direct pushes as the final
delivery unless the user explicitly asks for that exception.

---

## 2. Branch Rule

Use a focused branch for each activity or tightly related group of issues.

Branch names should describe the issue or scope, for example:
- `feature/24-27-e2e-ci-foundation`
- `docs/project-historical-resolution-prs`
- `fix/runtime-request-id-tests`

---

## 3. PR Body Rule

Every PR must include:
- a short summary of the delivered changes
- validation commands that were executed
- linked issues using `Closes #NN` when the PR is intended to close them
- any known limitation, blocked CI, or follow-up risk

---

## 4. Project Rule

When the issue is tracked in GitHub Projects:
- move the item to `In Progress` after opening the PR
- keep it out of `Done` until the PR is merged
- after merge, confirm the item has a linked PR before moving it to `Done`

---

## 5. Final Response Rule

When finishing an activity, report:
- PR number and URL
- linked issues
- validation results
- Project status updates, when applicable
