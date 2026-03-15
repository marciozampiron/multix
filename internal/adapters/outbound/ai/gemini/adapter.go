package gemini

import (
	"context"
	"fmt"
	"strings"
	"time"

	"multix/internal/domain/ai"
	"multix/internal/platform/logger"
	"multix/internal/ports/outbound"
)

type adapter struct {
	logger logger.Logger
}

func NewAdapter(log logger.Logger) outbound.AIProvider {
	return &adapter{
		logger: log.With("ai_provider", "gemini"),
	}
}

func (g *adapter) ID() string {
	return "gemini"
}

func (g *adapter) Generate(ctx context.Context, prompt ai.Prompt) (*ai.Response, error) {
	g.logger.Info("Sending prompt to Gemini API (Enterprise stub)", "length", len(prompt.Text))

	respText := fmt.Sprintf("Here is a detailed suggestion from Gemini for: '%s'", prompt.Text)
	return &ai.Response{
		Text:         respText,
		TokensUsed:   len(prompt.Text) / 4,
		ProviderName: g.ID(),
		GeneratedAt:  time.Now(),
	}, nil
}

func (g *adapter) SuggestCommand(ctx context.Context, intent string) (string, error) {
	g.logger.Info("Asking Gemini for CLI command suggestion", "intent", intent)

	intentLower := strings.ToLower(intent)
	if strings.Contains(intentLower, "eks") || strings.Contains(intentLower, "k8s") {
		return "multix k8s cluster list", nil
	}
	if strings.Contains(intentLower, "recurso") || strings.Contains(intentLower, "inventory") {
		return "multix inventory list --provider aws", nil
	}

	return "echo 'Try checking multix help'", nil
}
