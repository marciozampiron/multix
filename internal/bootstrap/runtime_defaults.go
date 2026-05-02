// File: internal/bootstrap/runtime_defaults.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 02/05/2026
// Purpose: Applies runtime configuration defaults to the CLI command tree.

package bootstrap

import (
	"fmt"
	"multix/internal/domain/config"

	"github.com/spf13/cobra"
)

const (
	runtimeProviderFlag = "provider"
	runtimeOutputFlag   = "output"
)

type runtimeFlagDefault struct {
	name  string
	value string
}

// ApplyRuntimeDefaults maps runtime configuration values onto root CLI defaults.
func ApplyRuntimeDefaults(rootCmd *cobra.Command, cfg *config.Config) error {
	if rootCmd == nil {
		return fmt.Errorf("root command is required")
	}
	if cfg == nil {
		return fmt.Errorf("runtime config is required")
	}

	defaults := []runtimeFlagDefault{
		{name: runtimeProviderFlag, value: cfg.DefaultCloudProvider},
		{name: runtimeOutputFlag, value: cfg.DefaultOutputMode},
	}

	for _, def := range defaults {
		if err := setPersistentFlagDefault(rootCmd, def.name, def.value); err != nil {
			return err
		}
	}

	return nil
}

func setPersistentFlagDefault(rootCmd *cobra.Command, name, value string) error {
	flags := rootCmd.PersistentFlags()
	flag := flags.Lookup(name)
	if flag == nil {
		return fmt.Errorf("persistent flag %q not found", name)
	}

	if err := flags.Set(name, value); err != nil {
		return fmt.Errorf("set persistent flag %q default: %w", name, err)
	}
	flag.DefValue = value

	return nil
}
