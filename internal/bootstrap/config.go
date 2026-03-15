// File: internal/bootstrap/config.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 15/03/2026
// Purpose: Bootstraps the in-memory application configuration logic.

package bootstrap

import (
	"multix/internal/domain/config"
	"os"
)

// LoadConfig creates a simple in-memory config bootstrap.
// Defaults are hardcoded, and can be overridden by environment variables.
func LoadConfig() *config.Config {
	cfg := &config.Config{
		AppName:              "multix",
		Version:              "0.2.0",
		DefaultCloudProvider: "aws",
		DefaultAIProvider:    "gemini",
		DefaultOutputMode:    "json",
	}

	if provider := os.Getenv("MULTIX_PROVIDER"); provider != "" {
		cfg.DefaultCloudProvider = provider
	}
	if out := os.Getenv("MULTIX_OUTPUT"); out != "" {
		cfg.DefaultOutputMode = out
	}
	if ai := os.Getenv("MULTIX_AI_PROVIDER"); ai != "" {
		cfg.DefaultAIProvider = ai
	}

	return cfg
}
