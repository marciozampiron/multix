# Review Comments — Multix

You are working inside the Multix repository.

Perform a **documentation review only**.

IMPORTANT:
- Do NOT modify code.
- Do NOT rewrite files.
- Do NOT implement changes.
- This is a REVIEW-ONLY pass.
- Respect GEMINI.md and all .agent/*.md rules.
- Respect docs/standards.

## REVIEW GOALS

Review the repository for documentation quality and consistency:

1. Missing file headers in important Go files
2. Missing Go doc comments on exported symbols
3. Weak or noisy comments
4. Over-commented code
5. Missing architectural intent comments where helpful
6. Inconsistent header format
7. Comments that violate Go idioms
8. Comments that describe obvious mechanics instead of intent

## REQUIRED OUTPUT FORMAT

Return exactly:

### Documentation Review Summary
- Overall assessment: PASS / PASS WITH CHANGES / FAIL
- Short summary

### Findings
For each finding:
- Severity: HIGH / MEDIUM / LOW
- File
- Problem
- Why it violates repo documentation rules
- Recommended minimal fix

### Coverage Check
Explicitly confirm whether:
- important files have file headers
- exported symbols have Go doc comments
- comments are mostly idiomatic Go
- the codebase is ready for a documentation-only patch