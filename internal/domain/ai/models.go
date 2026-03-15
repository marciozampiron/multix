package ai

import "time"

// Prompt represents an input to an AI assistant.
type Prompt struct {
	Text      string
	Context   string
	CreatedAt time.Time
}

// Response represents the returned generation from an AI assistant.
type Response struct {
	Text         string
	TokensUsed   int
	ProviderName string
	GeneratedAt  time.Time
}
