# ADR 0002 — Provider Registry

## Status
Accepted

## Context
The platform needs to support multiple cloud and AI providers without leaking provider details into application logic.

## Decision
Providers will be resolved through explicit registries and ports.

## Consequences
- provider substitution becomes easier
- CLI remains stable
- skills remain more reusable