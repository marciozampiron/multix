# Comment Pass — Multix

You are working inside the Multix repository.

Apply a **documentation-only pass** to the codebase.

IMPORTANT:
- Do NOT change business logic.
- Do NOT refactor behavior.
- Do NOT change public contracts unless strictly necessary to add documentation.
- Do NOT rename packages, files, symbols, or imports.
- Do NOT move files.
- Do NOT add new abstractions.
- This is a COMMENT / DOCUMENTATION PASS ONLY.
- Respect GEMINI.md and all .agent/*.md rules.
- Respect docs/standards.

## OBJECTIVE

Improve documentation consistency by adding or fixing:

1. file headers for important Go files
2. Go doc comments for exported symbols
3. concise non-noisy comments only where they explain:
   - intent
   - tradeoffs
   - non-obvious constraints
   - architectural rationale

## MANDATORY RULES

### 1. File Headers
For important files, add this header format if missing:

```go
// File: internal/application/example/example.go
// Company: Hassan
// Creator: Zamp
// Created: DD/MM/YYYY
// Updated: DD/MM/YYYY
// Purpose: Short, clear description of the file's role in the ecosystem.
```

Apply headers mainly to:
- `cmd/**/*.go`
- `internal/application/**/*.go`
- `internal/ports/**/*.go`
- `internal/adapters/**/*.go`
- `internal/bootstrap/**/*.go`

Skip headers for:
- generated files
- trivial tiny files
- tiny tests with no meaningful business logic

### 2. Go Doc Comments

Add missing Go doc comments for important exported symbols:
- exported interfaces
- exported structs
- exported constructors
- exported functions
- exported methods that define stable contracts

Rules:
- start with the symbol name
- keep the first sentence concise
- explain the role/contract
- keep it go doc friendly

### 3. Comment Quality

Only add comments that explain:
- why something exists
- non-obvious constraints
- tradeoffs
- architectural intent

Do NOT add noisy comments like:
- increment counter
- set variable
- call function
- return result

## EXECUTION MODE

Inspect the current files first.
Make the smallest documentation-only changes possible.
Prefer editing existing files only.
Do not change behavior.
If a change would alter logic, skip it and report it instead.

## REQUIRED OUTPUT FORMAT

Return exactly:

### Documentation Pass Summary
- Files changed
- Files skipped (if any)
- What was added:
  - file headers
  - Go doc comments
  - inline rationale comments (if any)

### Updated Code
Provide the updated code for each changed file

### Sanity Check
Explicitly confirm:
- no business logic changed
- no signatures changed (unless absolutely required for docs formatting, which should be avoided)
- no packages renamed
- no files moved
- no new abstractions introduced
- only documentation/comment changes were made
