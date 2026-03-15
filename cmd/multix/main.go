package main

import (
	"fmt"
	"os"

	"multix/internal/bootstrap"
)

func main() {
	app := bootstrap.BuildApp()
	cmd := bootstrap.Wire(app)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime Error: %v\n", err)
		os.Exit(1)
	}
}
