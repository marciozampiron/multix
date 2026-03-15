package main

import (
	"fmt"
	"os"

	"multix/internal/bootstrap"
)

func main() {
	container := bootstrap.BuildContainer()
	rootCmd := bootstrap.Wire(container)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime Error: %v\n", err)
		os.Exit(1)
	}
}
