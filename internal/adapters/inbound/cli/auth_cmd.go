package cli

import (
	"multix/internal/application/skills"

	"github.com/spf13/cobra"
)

type authHandler struct {
	rootCmd  *cobra.Command
	executor *skills.Executor
}

func NewAuthHandler(root *cobra.Command, executor *skills.Executor) *authHandler {
	return &authHandler{rootCmd: root, executor: executor}
}

func (h *authHandler) Register() {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication and access validation across providers",
	}

	validateCmd := &cobra.Command{
		Use:           "validate",
		Short:         "Validate current provider credentials",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider, _ := cmd.Flags().GetString("provider")
			outFmt, _ := cmd.Flags().GetString("output")

			res, err := h.executor.Execute(cmd.Context(), "auth.validate", map[string]any{"provider": provider})
			if err != nil {
				return err
			}

			return render(res, outFmt)
		},
	}

	whoamiCmd := &cobra.Command{
		Use:           "whoami",
		Short:         "Display current active identity",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider, _ := cmd.Flags().GetString("provider")
			outFmt, _ := cmd.Flags().GetString("output")

			res, err := h.executor.Execute(cmd.Context(), "auth.whoami", map[string]any{"provider": provider})
			if err != nil {
				return err
			}

			return render(res, outFmt)
		},
	}

	authCmd.AddCommand(validateCmd)
	authCmd.AddCommand(whoamiCmd)
	h.rootCmd.AddCommand(authCmd)
}
