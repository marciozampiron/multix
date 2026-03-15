# Multix Prompt Library

This directory contains reusable prompt templates for Gemini CLI / Gemini Code Assist agent workflows.

## Primary workflow

1. `upgrade-v0.2-audit.md`
   - audit + plan only
2. `upgrade-v0.2-implement.md`
   - implementation only after approval
3. `review-diff.md`
   - review only, no code changes

## Additional prompts

- `safe-fix.md`
  - minimal safe patch for small fixes
- `feature-template.md`
  - reusable template for new feature work

## Rules

- Prefer patch-style evolution
- Prefer plan before implementation
- Prefer review before merge
- Respect `GEMINI.md` and `.agent/*.md`
