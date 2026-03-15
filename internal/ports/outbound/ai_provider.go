package outbound

import (
	"context"

	"multix/internal/domain/ai"
)

// AIProvider defines the methods needed for a Generative AI backend.
type AIProvider interface {
	// ID returns the unique identifier for this AI provider.
	ID() string

	// Generate generates a rich text response given a prompt.
	Generate(ctx context.Context, prompt ai.Prompt) (*ai.Response, error)

	// SuggestCommand asks the AI to recommend a CLI command based on an intent string.
	SuggestCommand(ctx context.Context, intent string) (string, error)
}
