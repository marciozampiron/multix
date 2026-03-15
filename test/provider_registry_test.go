// File: test/provider_registry_test.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 15/03/2026
// Purpose: Validates provider registry resolution across supported capabilities.

package test

import (
	"testing"

	"multix/internal/bootstrap"
	"multix/internal/platform/logger"
)

func TestProviderRegistryResolution(t *testing.T) {
	log := logger.New("none")
	providers := bootstrap.BuildProviderRegistry(log)

	t.Run("resolves known providers", func(t *testing.T) {
		tests := []struct {
			name    string
			resolve func() error
		}{
			{
				name: "aws auth",
				resolve: func() error {
					p, err := providers.GetCloudAuthProvider("aws")
					if err != nil {
						return err
					}
					if p == nil {
						t.Fatal("expected AWS auth provider, got nil")
					}
					return nil
				},
			},
			{
				name: "gcp inventory",
				resolve: func() error {
					p, err := providers.GetCloudInventoryProvider("gcp")
					if err != nil {
						return err
					}
					if p == nil {
						t.Fatal("expected GCP inventory provider, got nil")
					}
					return nil
				},
			},
			{
				name: "aws k8s",
				resolve: func() error {
					p, err := providers.GetKubernetesProvider("aws")
					if err != nil {
						return err
					}
					if p == nil {
						t.Fatal("expected AWS k8s provider, got nil")
					}
					return nil
				},
			},
			{
				name: "gemini ai",
				resolve: func() error {
					p, err := providers.GetAIProvider("gemini")
					if err != nil {
						return err
					}
					if p == nil {
						t.Fatal("expected Gemini AI provider, got nil")
					}
					return nil
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if err := tt.resolve(); err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			})
		}
	})

	t.Run("returns error for unknown providers", func(t *testing.T) {
		tests := []struct {
			name    string
			resolve func() error
		}{
			{
				name: "unknown auth",
				resolve: func() error {
					_, err := providers.GetCloudAuthProvider("unknown-cloud")
					return err
				},
			},
			{
				name: "unknown inventory",
				resolve: func() error {
					_, err := providers.GetCloudInventoryProvider("unknown-cloud")
					return err
				},
			},
			{
				name: "unknown k8s",
				resolve: func() error {
					_, err := providers.GetKubernetesProvider("unknown-cloud")
					return err
				},
			},
			{
				name: "unknown ai",
				resolve: func() error {
					_, err := providers.GetAIProvider("unknown-ai")
					return err
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if err := tt.resolve(); err == nil {
					t.Fatal("expected error for unknown provider, got nil")
				}
			})
		}
	})
}

func TestProviderRegistryResolution_NormalizesProviderNames(t *testing.T) {
	log := logger.New("none")
	providers := bootstrap.BuildProviderRegistry(log)

	tests := []struct {
		name string
	}{
		{name: "AWS"},
		{name: " aws "},
		{name: "GCP"},
		{name: " Gemini "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "AWS", " aws ":
				p, err := providers.GetCloudAuthProvider(tt.name)
				if err != nil || p == nil {
					t.Fatalf("expected normalized auth provider lookup to succeed for %q", tt.name)
				}
			case "GCP":
				p, err := providers.GetCloudInventoryProvider(tt.name)
				if err != nil || p == nil {
					t.Fatalf("expected normalized inventory provider lookup to succeed for %q", tt.name)
				}
			case " Gemini ":
				p, err := providers.GetAIProvider(tt.name)
				if err != nil || p == nil {
					t.Fatalf("expected normalized ai provider lookup to succeed for %q", tt.name)
				}
			}
		})
	}
}
