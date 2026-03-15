You are working inside the Multix repository.

We are upgrading the current scaffold from v0.1 to v0.2.

IMPORTANT:
- This is a controlled in-place upgrade.
- Do NOT rewrite the repository.
- Do NOT implement code yet.
- Do NOT create files yet.
- Do NOT edit files yet.
- Respect GEMINI.md and all .agent/*.md rules.
- Respect docs/standards and documentation-in-code rules.
- This step is AUDIT + PLAN ONLY.

## MANDATORY EXECUTION MODE

Before any implementation, you MUST:

1. Inspect the current repository structure.
2. Identify which v0.1 elements already exist.
3. Reuse existing files and abstractions whenever possible.
4. Detect the smallest possible delta to reach v0.2.
5. Prefer modifying existing files over adding new files.
6. Avoid moving files unless strictly necessary.
7. Avoid renaming packages unless strictly necessary.
8. Avoid creating duplicate registries, duplicate skill contracts, duplicate command trees, or duplicate bootstrap flows.
9. Treat this as a patch-style evolution, not a redesign.
10. If the actual repository differs materially from the expected v0.1 scaffold, STOP and say:
   "Repository differs from expected baseline. I need to adjust the plan to the actual structure before implementation."

## REQUIRED OUTPUT FORMAT

Return your answer in exactly these sections:

### Phase 1 — Repository Audit
- List relevant existing files found
- Summarize what is already implemented from v0.1
- Identify what is missing for v0.2
- Highlight any risk areas, ambiguities, or possible conflicts

### Phase 2 — Minimal Change Plan
- List files to change
- List files to add (ONLY if truly necessary)
- For each file, explain WHY it must change
- Explicitly justify every new file
- If a new file is optional, prefer NOT creating it

### Phase 3 — Safety Review
- Confirm whether the plan avoids:
  - duplicate registries
  - duplicate skill contracts
  - duplicate command trees
  - duplicate bootstrap paths
  - unnecessary refactors
- Call out any place where the current repo structure should be preserved exactly

## Final instruction
For THIS response:
- Execute ONLY Phase 1, Phase 2, and Phase 3.
- Do NOT provide code.
- Do NOT implement yet.
- Wait for explicit approval before implementation.
