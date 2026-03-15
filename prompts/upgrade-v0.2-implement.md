Approved.

Now execute ONLY the implementation based on the previously approved minimal change plan.

IMPORTANT:
- This is a controlled patch-style implementation.
- Do NOT expand scope.
- Do NOT redesign the architecture.
- Do NOT introduce new abstractions unless they were explicitly justified in the approved plan.
- Respect GEMINI.md and all .agent/*.md rules.
- Respect docs/standards and documentation-in-code rules.
- Preserve the current scaffold structure.
- Preserve package names and import paths whenever possible.
- Prefer modifying existing files over creating new files.

## IMPLEMENTATION RULES

1. Implement ONLY the approved minimal set of changes.
2. Do NOT create additional files beyond the approved list.
3. Do NOT move files unless explicitly approved.
4. Do NOT rename packages unless explicitly approved.
5. Do NOT create duplicate registries.
6. Do NOT create duplicate skill abstractions.
7. Do NOT create parallel command trees.
8. Do NOT create parallel bootstrap flows.
9. Keep diffs small and coherent.
10. Follow file header and Go doc comment rules for all important files.
11. If the real code structure differs from the approved plan, STOP and explain the mismatch before continuing.

## REQUIRED OUTPUT FORMAT

Return your answer in exactly these sections:

### Phase 3 — Implementation
- Files changed
- Files added
- Updated code for each changed file
- Full code for each new file
- Short rationale for each file

### Phase 4 — Sanity Check
Explicitly confirm:
- no duplicate registries were created
- no duplicate skill contracts were created
- no duplicate command trees were introduced
- no parallel bootstrap paths were introduced
- no unnecessary refactor was done
- the repository still follows skills-first + agent-ready architecture
- file headers were applied where required
- Go doc comments were applied where required

## Final instruction
Do NOT propose extra enhancements.
Do NOT suggest future improvements in this response.
Only implement the approved patch.
