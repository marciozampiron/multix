package root

import (
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "multix",
		Short: "Multix Enterprise CLI - Skills First",
		Long:  `A fast, flexible, and capable CLI built on a Skills-First / DDD architecture.`,
	}

	rootCmd.PersistentFlags().StringP("log-level", "l", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringP("provider", "p", "aws", "Target implementation provider")
	rootCmd.PersistentFlags().StringP("output", "o", "table", "Output format (json, table)")

	return rootCmd
}

func RegisterVersionCmd(rootCmd *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the application version",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Multix Enterprise CLI: v1.1.0-skills-first")
		},
	}
	rootCmd.AddCommand(cmd)
}
