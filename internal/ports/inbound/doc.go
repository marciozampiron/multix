// Package inbound defines the v1 inbound port contracts that drive the
// Multix runtime from the outside (CLI handlers and AI agent tool callers).
//
// # Stability promise
//
// Same semver rules as ports/outbound apply: methods on these interfaces
// are part of the public surface once Multix reaches v1.0. Renames or
// signature changes require a major-version bump of the package path.
package inbound
