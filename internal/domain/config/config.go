// File: internal/domain/config/config.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 15/03/2026
// Purpose: Defines the centralized configuration models.

package config

// Config holds the runtime configuration parameters for the Multix platform.
type Config struct {
	AppName              string
	Version              string
	DefaultCloudProvider string
	DefaultAIProvider    string
	DefaultOutputMode    string
}
