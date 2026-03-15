// File: cmd/multix/main.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 15/03/2026
// Purpose: Executable binary entrypoint for the Multix CLI ecosystem.

package main

import (
	"fmt"
	"os"

	"multix/internal/bootstrap"
)

func main() {
	app, err := bootstrap.BuildApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Bootstrap Error: %v\n", err)
		os.Exit(1)
	}

	cmd := app.Wire()

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime Error: %v\n", err)
		os.Exit(1)
	}
}
