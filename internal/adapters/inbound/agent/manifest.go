// File: internal/adapters/inbound/agent/manifest.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 15/03/2026
// Purpose: Generates provider-agnostic tool manifests from registered skills.

package agent

import (
	"sort"

	domainSkills "multix/internal/domain/skills"
)

// Manifest represents a provider-agnostic tool manifest derived from a registered skill.
type Manifest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

// GenerateManifests extracts all registered skills and formats them as tool manifests.
func GenerateManifests(registry *domainSkills.Registry) []Manifest {
	if registry == nil {
		return []Manifest{}
	}

	var manifests []Manifest

	for _, skill := range registry.ListAll() {
		manifests = append(manifests, Manifest{
			Name:        skill.Name(),
			Description: skill.Description(),
			Parameters:  skill.InputSchema(),
		})
	}

	sort.Slice(manifests, func(i, j int) bool {
		return manifests[i].Name < manifests[j].Name
	})

	return manifests
}
