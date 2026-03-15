# Go File Template

Use this template for important new files.

```go
// File: internal/application/example/example.go
// Company: Hassan
// Creator: Zamp
// Created: DD/MM/YYYY
// Updated: DD/MM/YYYY
// Purpose: Short, clear description of the file's role in the ecosystem.

package example

import (
	"context"
)

// ExampleUseCase executes the example flow for the target capability.
type ExampleUseCase struct {
	// dependencies
}

// NewExampleUseCase creates a new ExampleUseCase.
func NewExampleUseCase() *ExampleUseCase {
	return &ExampleUseCase{}
}

// Execute runs the example use case.
func (u *ExampleUseCase) Execute(ctx context.Context) error {
	return nil
}