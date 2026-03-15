package cli

import (
	"context"
	"fmt"

	"multix/internal/application/skills"

	"github.com/spf13/cobra"
)

type doctorHandler struct {
	rootCmd  *cobra.Command
	executor *skills.Executor
}

func NewDoctorHandler(root *cobra.Command, executor *skills.Executor) *doctorHandler {
	return &doctorHandler{rootCmd: root, executor: executor}
}

func (h *doctorHandler) Register() {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check system readiness and dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := h.executor.Execute(context.Background(), "doctor.run", map[string]any{})
			if err != nil {
				return err
			}

			payload := res.(map[string]any)
			checks := payload["checks"].([]map[string]any)
			fmt.Printf("Running diagnostic checks...\n")
			for _, check := range checks {
				if check["ok"].(bool) {
					fmt.Printf("✔ %s OK\n", check["name"])
				} else {
					fmt.Printf("✖ %s MISSING\n", check["name"])
				}
			}
			return nil
		},
	}
	h.rootCmd.AddCommand(cmd)
}
