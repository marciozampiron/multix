// Package outbound defines the v1 outbound port contracts that adapters must
// satisfy to participate in the Multix runtime.
//
// # Stability promise
//
// Once Multix reaches v1.0, every interface in this package is part of the
// public extension surface and follows semver:
//
//   - Methods are NEVER renamed, removed or have their signatures changed in
//     a minor release. Such changes require a new major version of the
//     module path (e.g. .../ports/outbound/v2).
//   - New methods MAY be added in a minor release ONLY when accompanied by
//     a default helper or embedded "Plus" interface so existing
//     implementations continue to compile. Embedding the v1 interface in a
//     v1Plus interface is the preferred pattern.
//   - Method comments and behavioral semantics are part of the contract:
//     adapters that violate them are considered non-conformant.
//
// # Contract testing
//
// Each adapter must declare compile-time interface compliance via
//
//	var _ outbound.AuthProvider = (*Adapter)(nil)
//
// and ship a contract test under its package. The shared
// `outbound/contracttest` helpers (when introduced) will exercise the
// behavioral semantics generically.
//
// # Conventions
//
//   - All methods MUST accept a context.Context as their first argument
//     (except ID()).
//   - Implementations MUST be safe for concurrent use unless explicitly
//     documented otherwise.
//   - Errors returned MUST wrap the underlying SDK error with %w when one
//     exists, and MUST be actionable (mention the env var, command or
//     remediation step).
package outbound
