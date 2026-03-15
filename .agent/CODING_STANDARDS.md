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

---

## 3. Function Design

- Keep functions focused
- Keep functions small when practical
- Prefer one clear responsibility per function
- Avoid deep nesting
- Return early on errors
- Do not create "god functions"

---

## 4. Comments Policy

### 4.1 Comment only when useful
Comments are encouraged only when they explain:
- why something exists
- non-obvious constraints
- tradeoffs
- architectural decisions
- exported APIs (when useful)

### 4.2 Avoid noise
Do NOT add comments like:
- "increment i"
- "call function"
- "set variable"
- "return result"

---

## 5. File Headers (Mandatory for Key Files)

Important files must include a short standardized header at the top.

Recommended format:

```go
// File: internal/application/inventory/scan.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 15/03/2026
// Purpose: Executes the inventory scan use case for reusable platform and agent skill flows.