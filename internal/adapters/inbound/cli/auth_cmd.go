package cli

import (
	"context"
	"fmt"

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

	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Login to a provider",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := h.executor.Execute(context.Background(), "auth.login", map[string]any{"provider": "aws"})
			if err != nil {
				return err
			}

			payload := res.(map[string]any)
			fmt.Printf("Successfully logged in to %s (valid: %v)\n", payload["provider"], payload["valid"])
			return nil
		},
	}

	whoamiCmd := &cobra.Command{
		Use:   "whoami",
		Short: "Display current active identity",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := h.executor.Execute(context.Background(), "auth.whoami", map[string]any{"provider": "aws"})
			if err != nil {
				return err
			}

			payload := res.(map[string]any)
			fmt.Printf("Current Identity: %s under account %s\n", payload["username"], payload["account_id"])
			return nil
		},
	}

	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(whoamiCmd)
	h.rootCmd.AddCommand(authCmd)
}
