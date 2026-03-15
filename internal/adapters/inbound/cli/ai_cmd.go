package cli

import (
	"context"
	"github.com/spf13/cobra"
	"multix/internal/application/skills"
)

type aiHandler struct {
	rootCmd  *cobra.Command
	executor *skills.Executor
}

func NewAIHandler(root *cobra.Command, executor *skills.Executor) *aiHandler {
	return &aiHandler{rootCmd: root, executor: executor}
}

func (h *aiHandler) Register() {
	cmd := &cobra.Command{
		Use:   "ai",
		Short: "Use Generative AI to accelerate DevOps tasks",
	}

	explainCmd := &cobra.Command{
		Use:   "explain [prompt]",
		Short: "Ask the AI assistant to explain a resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := "gemini" // Default MVP provider
			outFmt, _ := cmd.Flags().GetString("output")
			res, err := h.executor.Execute(context.Background(), "ai.explain", map[string]any{"provider": provider, "prompt": args[0]})
			if err != nil {
				return err
			}

			return render(res, outFmt)
		},
	}

	cmd.AddCommand(explainCmd)
	h.rootCmd.AddCommand(cmd)
}
