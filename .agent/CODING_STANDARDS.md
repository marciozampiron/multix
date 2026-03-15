# Multix Coding Standards

## 1. General Principles

- Write idiomatic Go
- Prefer simplicity over cleverness
- Keep code explicit
- Minimize hidden magic
- Optimize for readability and maintainability
- Favor stable, boring, production-friendly patterns

---

## 2. Code Style

- Use `gofmt`
- Keep imports organized
- Prefer short package names
- Prefer descriptive local names when needed
- Avoid stuttered names

Good:
- `registry`
- `executor`
- `provider`
- `cfg`

Bad:
- `skillRegistryManagerService`
- `providerImplementationFactoryManager`

---

## 3. Function Design

- Keep functions focused
- Keep functions small when practical
- Prefer one clear responsibility per function
- Avoid deep nesting
- Return early on errors
- Do not create "god functions"

Good:
- parse → validate → execute → return

Bad:
- giant handlers mixing CLI, business logic, and formatting

---

## 4. Comments Policy

## 4.1 Comment only when useful
Comments are encouraged only when they explain:
- why something exists
- non-obvious constraints
- tradeoffs
- architectural decisions
- exported APIs (when useful)

## 4.2 Avoid noise
Do NOT add comments like:
- "increment i"
- "call function"
- "set variable"
- "return result"

## 4.3 Exported symbols
If a symbol is exported and part of a stable contract, prefer concise doc comments.

---

## 5. Interfaces

- Do not create interfaces before they are needed
- Keep interfaces small
- Define interfaces near consumers when practical
- Avoid "service" abstractions with no benefit
- Prefer concrete types internally unless substitution is actually needed

Good:
- `AIProvider`
- `CloudAuthProvider`

Bad:
- `ProviderManagerServiceFactoryRepository`

---

## 6. Error Handling

- Always return errors explicitly
- Never swallow errors silently
- Wrap errors when crossing layers if it adds context
- Keep error messages actionable
- Avoid panic except for truly unrecoverable programmer errors

Good:
- `return nil, fmt.Errorf("validate credentials: %w", err)`

Bad:
- `return nil, err` everywhere without context across layers
- `panic(err)` in application flow

---

## 7. Logging

- Use structured logging
- Prefer `log/slog`
- Keep logs concise
- Never log secrets
- Avoid excessive logs in hot paths
- Use logs for events and failures, not for narrating obvious execution

---

## 8. Context Usage

Use `context.Context` in:
- provider calls
- I/O flows
- network-related operations
- skill execution

Do not:
- store context in structs
- pass nil contexts
- ignore cancellation if it matters

---

## 9. Performance Rules

- Avoid unnecessary allocations
- Avoid reflection unless clearly justified
- Avoid converting large objects repeatedly
- Keep hot paths simple
- Keep startup path lightweight
- Do not initialize all providers eagerly if avoidable

---

## 10. Dependency Rules

Before adding a dependency:
1. Is stdlib enough?
2. Is the dependency stable and widely used?
3. Does it add real value?
4. Is there a lighter alternative?
5. Does it increase startup cost or complexity?

MVP preference:
- stdlib first
- Cobra where needed
- minimal external dependencies

---

## 11. Formatting & Linting

Always run before finalizing:
- `make fmt`
- `make vet`
- `make test`

If available:
- `make lint`
- `make vuln`