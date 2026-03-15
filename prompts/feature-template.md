You are working inside the Multix repository.

Implement a NEW FEATURE using patch-style evolution.

## Workflow
1. Audit current repo
2. Propose minimal change plan
3. Wait for approval
4. Implement only approved patch
5. Review for overengineering

## Architecture rules
- skills-first
- agent-ready
- thin CLI handlers
- business logic outside Cobra
- explicit wiring
- provider abstractions behind ports
- no duplicate registries
- no duplicate command trees
- no parallel bootstrap paths

## Required output
For the planning step:
- audit
- files to change
- files to add
- rationale
- safety review

For the implementation step:
- files changed
- files added
- updated code
- rationale
- sanity check
