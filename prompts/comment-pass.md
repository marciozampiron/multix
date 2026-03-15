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