// File: internal/application/doctor/doctor_skill.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 15/03/2026
// Purpose: Validates local execution dependencies and environment readiness.

package doctor

import (
	"context"

	"multix/internal/domain/skills"
)

// CheckEnvSkill is the universal tool for validating environment dependencies.
type CheckEnvSkill struct{}

// NewCheckEnvSkill creates a new CheckEnvSkill.
func NewCheckEnvSkill() skills.Skill {
	return &CheckEnvSkill{}
}

func (s *CheckEnvSkill) Name() string {
	return "doctor.run"
}

func (s *CheckEnvSkill) Description() string {
	return "Verifies the system environment dependencies such as Go, AWS CLI, kubectl."
}

func (s *CheckEnvSkill) InputSchema() any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

func (s *CheckEnvSkill) Execute(ctx context.Context, input map[string]any) (any, error) {
	// Dummy logic mimicking system health checks
	results := []map[string]any{
		{"name": "Go", "ok": true},
		{"name": "AWS CLI", "ok": true},
		{"name": "gcloud CLI", "ok": false},
	}
	return map[string]any{"checks": results}, nil
}
