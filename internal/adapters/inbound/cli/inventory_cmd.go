package cli

import (
	"context"
	"github.com/spf13/cobra"
	"multix/internal/application/skills"
	"multix/internal/platform/formatter"
)

type inventoryHandler struct {
	rootCmd  *cobra.Command
	executor *skills.Executor
}

func NewInventoryHandler(root *cobra.Command, executor *skills.Executor) *inventoryHandler {
	return &inventoryHandler{rootCmd: root, executor: executor}
}

func (h *inventoryHandler) Register() {
	cmd := &cobra.Command{
		Use:   "inventory",
		Short: "Discover and list multi-cloud resources",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List resources matching a specific type",
		RunE: func(cmd *cobra.Command, args []string) error {
			service, _ := cmd.Flags().GetString("service")
			provider, _ := cmd.Flags().GetString("provider")
			outFmt, _ := cmd.Flags().GetString("output")
			if provider == "" {
				provider = "aws"
			}

			// Passing JSON-like map, exactly as an Agent LLM would!
			input := map[string]any{"provider": provider, "service": service}

			res, err := h.executor.Execute(context.Background(), "inventory.scan", input)
			if err != nil {
				return err
			}

			return formatter.Print(res, formatter.OutputFormat(outFmt))
		},
	}
	listCmd.Flags().StringP("service", "s", "", "Filter by service/resource type (e.g., compute, ec2)")

	cmd.AddCommand(listCmd)
	h.rootCmd.AddCommand(cmd)
}
