# Multix Safety & Security Rules

## 1. Secrets

Never:
- hardcode credentials
- commit tokens
- print secrets
- log secrets
- embed API keys in examples

Use placeholders like:
- `YOUR_API_KEY`
- `MULTIX_API_TOKEN`

---

## 2. Logging Safety

When logging:
- never print tokens
- never print access keys
- never print secrets from config
- mask sensitive identifiers if needed

---

## 3. Shell Safety

Avoid shell execution unless explicitly justified.

If shell is needed:
- keep it explicit
- validate inputs
- avoid command injection risks

---

## 4. Provider Safety

When integrating real providers:
- validate credentials early
- use least privilege assumptions
- do not leak sensitive provider errors directly if they contain secrets

---

## 5. Dependency Safety

Before adding a dependency:
- justify necessity
- prefer stdlib first
- prefer mature libraries
- run vulnerability checks

Preferred command:
- `make vuln`

---

## 6. Plugin Safety

MVP plugin system must NOT use:
- native Go `.so` plugin loading

Prefer:
- manifest-based plugins
- external executables
- controlled adapters

---

## 7. Future AI Safety

When exposing skills to agents:
- keep input schemas constrained
- validate required fields
- keep outputs machine-friendly
- do not let prompt text become business logic
- do not let the agent bypass validation rules