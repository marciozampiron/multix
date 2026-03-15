Review the proposed implementation diff ONLY.

IMPORTANT:
- Do NOT modify code.
- Do NOT rewrite files.
- Do NOT implement anything.
- This is a review-only pass.
- Respect GEMINI.md and all .agent/*.md rules.
- Respect docs/standards and documentation-in-code rules.

## REVIEW GOALS

Review the proposed patch for:

1. Overengineering
2. Duplicate abstractions
3. Duplicate registries
4. Duplicate skill contracts
5. Parallel command trees
6. Parallel bootstrap flows
7. Unnecessary file creation
8. Unnecessary package churn
9. Violations of skills-first architecture
10. Business logic leaking into Cobra handlers
11. Missing file headers where required
12. Missing Go doc comments where required
13. Provider-specific leakage into application layer
14. Rendering logic leaking into skills/use cases
15. Test gaps for changed behavior

## REQUIRED OUTPUT FORMAT

Return exactly these sections:

### Review Summary
- Overall assessment: PASS / PASS WITH CHANGES / FAIL
- Short summary of quality and risk

### Findings
For each finding:
- Severity: HIGH / MEDIUM / LOW
- File
- Problem
- Why it violates the repo rules
- Recommended minimal fix

### Safe-to-Merge Check
Explicitly confirm whether:
- the patch is safe to merge as-is
- the patch needs small corrections
- the patch should be reworked before merge

## Final instruction
Do NOT provide rewritten code unless explicitly asked later.
This is review-only.
