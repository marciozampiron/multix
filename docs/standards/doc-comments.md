# Go Doc Comments Standard

## Objective

Define how exported symbols must be documented in the Multix repository.

This standard follows idiomatic Go practices and improves:
- `go doc` compatibility
- developer experience
- AI-assisted code understanding
- contributor onboarding

---

## Mandatory Rule

All important exported symbols must have Go doc comments.

This includes:
- exported interfaces
- exported structs
- exported constructors
- exported functions
- exported methods that define stable contracts

---

## Rules

- Start the comment with the symbol name
- Keep the first sentence concise
- Explain the role or contract
- Do not describe obvious mechanics
- Keep comments `go doc` friendly

---

## Good Example

```go
// ProviderRegistry resolves cloud and AI providers by stable logical name.
type ProviderRegistry struct {
	// ...
}

// NewProviderRegistry creates a new ProviderRegistry with registered providers.
func NewProviderRegistry() *ProviderRegistry {
	// ...
}