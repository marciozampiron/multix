# ADR 0007 — Tool Manifest Shape for Agent Tool-Calling

## Status
Accepted

## Context
LLM tool-calling APIs (Anthropic, OpenAI, Gemini) all consume a list of
tool descriptors that share a common subset: a name, a natural-language
description, and a JSON Schema for arguments. They differ in how the
descriptor is wrapped (`tools[]` vs `function_declarations[]`) and in
required schema dialect.

Multix needs ONE manifest shape that survives the round-trip to all
three providers. Diverging by provider would mean per-provider skill
manifests — explosion of code and divergence risk.

## Decision
The agent ToolAdapter renders every Skill into a single canonical
descriptor:

```json
{
  "name": "<domain>.<verb>",
  "description": "<sentence the LLM uses to decide when to call>",
  "input_schema": { "type": "object", "properties": { ... }, "required": [...] }
}
```

- `name` is the Skill's `Name()` value, locked by ADR 0001 and the v1.0
  contract (#19).
- `description` is the Skill's `Description()`.
- `input_schema` is whatever `InputSchema()` returns, validated by the
  contract test to marshal to a top-level JSON object with `type:
  "object"`.

Provider-specific shims are responsible for renaming `input_schema` to
`parameters` (OpenAI/Gemini) or wrapping in `function_declarations`
(Gemini specifically). Those shims live OUTSIDE the runtime — the
runtime always speaks the canonical shape on `GET /tools`.

## Consequences
- One source of truth for skill metadata; new skills are exposed to
  every supported LLM the moment they register.
- Schema is JSON Schema 2020-12-shaped, which all three vendors accept
  with at most a key-rename.
- Agents that don't tolerate the canonical shape need a thin adapter
  layer; that adapter is owned by the agent integrator, not Multix.
- Adding a new field to the manifest (e.g. `since_version`) is a
  non-breaking change because clients ignore unknown keys.
