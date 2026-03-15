package agent

import (
	domainSkills "multix/internal/domain/skills"
)

// Manifest represents an AI tool manifest ready to be consumed by OpenAI/Gemini/MCP.
type Manifest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

// GenerateManifests extracts all skills from the registry and formats them as tool manifests.
func GenerateManifests(registry *domainSkills.Registry) []Manifest {
	var manifests []Manifest

	for _, skill := range registry.ListAll() {
		manifests = append(manifests, Manifest{
			Name:        skill.Name(),
			Description: skill.Description(),
			Parameters:  skill.InputSchema(),
		})
	}
	return manifests
}
