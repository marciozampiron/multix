// Package skills hosts the universal Skill contract — the single interface
// that every Multix capability must satisfy to be exposed simultaneously to
// the CLI, the runtime HTTP surface and the AI agent tool-calling protocol.
//
// # Stability promise
//
// The Skill interface is the single most important contract in Multix. It
// follows the same semver rules as ports/* and is locked at v1.0:
//
//   - The four methods (Name, Description, InputSchema, Execute) and their
//     signatures are frozen.
//   - Skill names use the `<domain>.<verb>` convention (e.g. `inventory.scan`,
//     `doctor.auth`). A name is a stable identifier — once shipped, it is
//     never repurposed for a different capability.
//   - InputSchema MUST return a value that marshals to a valid JSON Schema
//     draft 2020-12 object. The schema is the contract — tightening it
//     (forbidding previously accepted input) is a breaking change.
//   - Execute MUST be safe for concurrent invocation.
package skills
