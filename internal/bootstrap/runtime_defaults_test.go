package bootstrap

import (
	"strings"
	"testing"

	"multix/cmd/root"
	"multix/internal/domain/config"

	"github.com/spf13/cobra"
)

func TestApplyRuntimeDefaults(t *testing.T) {
	rootCmd := root.NewRootCmd()
	cfg := &config.Config{
		DefaultCloudProvider: "oci",
		DefaultOutputMode:    "table",
	}

	if err := ApplyRuntimeDefaults(rootCmd, cfg); err != nil {
		t.Fatalf("unexpected error applying runtime defaults: %v", err)
	}

	assertFlagValue(t, rootCmd, runtimeProviderFlag, "oci")
	assertFlagValue(t, rootCmd, runtimeOutputFlag, "table")
	assertFlagDefaultValue(t, rootCmd, runtimeProviderFlag, "oci")
	assertFlagDefaultValue(t, rootCmd, runtimeOutputFlag, "table")
}

func TestApplyRuntimeDefaultsRequiresRootCommand(t *testing.T) {
	cfg := &config.Config{
		DefaultCloudProvider: "aws",
		DefaultOutputMode:    "json",
	}

	err := ApplyRuntimeDefaults(nil, cfg)
	if err == nil || !strings.Contains(err.Error(), "root command is required") {
		t.Fatalf("expected root command error, got %v", err)
	}
}

func TestApplyRuntimeDefaultsRequiresConfig(t *testing.T) {
	err := ApplyRuntimeDefaults(root.NewRootCmd(), nil)
	if err == nil || !strings.Contains(err.Error(), "runtime config is required") {
		t.Fatalf("expected runtime config error, got %v", err)
	}
}

func TestApplyRuntimeDefaultsRequiresExpectedFlags(t *testing.T) {
	rootCmd := &cobra.Command{Use: "multix"}
	cfg := &config.Config{
		DefaultCloudProvider: "aws",
		DefaultOutputMode:    "json",
	}

	err := ApplyRuntimeDefaults(rootCmd, cfg)
	if err == nil || !strings.Contains(err.Error(), `persistent flag "provider" not found`) {
		t.Fatalf("expected missing provider flag error, got %v", err)
	}
}

func assertFlagValue(t *testing.T, rootCmd *cobra.Command, name, want string) {
	t.Helper()

	got, err := rootCmd.PersistentFlags().GetString(name)
	if err != nil {
		t.Fatalf("expected flag %q to exist: %v", name, err)
	}
	if got != want {
		t.Fatalf("expected flag %q value %q, got %q", name, want, got)
	}
}

func assertFlagDefaultValue(t *testing.T, rootCmd *cobra.Command, name, want string) {
	t.Helper()

	flag := rootCmd.PersistentFlags().Lookup(name)
	if flag == nil {
		t.Fatalf("expected flag %q to exist", name)
	}
	if flag.DefValue != want {
		t.Fatalf("expected flag %q default %q, got %q", name, want, flag.DefValue)
	}
}
