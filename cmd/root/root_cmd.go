package root

import (
	"multix/pkg/version"

	"github.com/spf13/cobra"
)

// NewRootCmd creates the root command for the Multix CLI.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "multix",
		Short:         "Multix Enterprise CLI - Skills First",
		Long:          `A fast, flexible, and capable CLI built on a Skills-First / DDD architecture.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.PersistentFlags().StringP("log-level", "l", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringP("provider", "p", "aws", "Target implementation provider")
	rootCmd.PersistentFlags().StringP("output", "o", "json", "Output format (json, table)")

	return rootCmd
}

// RegisterVersionCmd registers the version command in the root command tree.
func RegisterVersionCmd(rootCmd *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the application version",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Multix Enterprise CLI:", version.Version)
		},
	}
	rootCmd.AddCommand(cmd)
}
