// File: internal/domain/skills/registry.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 15/03/2026
// Purpose: Implements the central runtime registry for available skills.

package skills

import (
	"fmt"
	"sort"
	"sync"
)

// Registry manages the collection of available Agent Skills.
type Registry struct {
	mu     sync.RWMutex
	skills map[string]Skill
}

// NewRegistry creates a new instance of the skill registry.
func NewRegistry() *Registry {
	return &Registry{
		skills: make(map[string]Skill),
	}
}

// Register adds a new skill to the registry.
func (r *Registry) Register(skill Skill) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.skills[skill.Name()] = skill
}

// Get finds a skill by its exact name.
func (r *Registry) Get(name string) (Skill, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skill, exists := r.skills[name]
	if !exists {
		return nil, fmt.Errorf("skill '%s' not found", name)
	}
	return skill, nil
}

// ListAll returns all registered skills sorted by name.
func (r *Registry) ListAll() []Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []Skill
	for _, sl := range r.skills {
		list = append(list, sl)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Name() < list[j].Name()
	})
	return list
}
