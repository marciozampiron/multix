package cli

import (
	"context"
	"fmt"

	"multix/internal/application/skills"

	"github.com/spf13/cobra"
)

type k8sHandler struct {
	rootCmd  *cobra.Command
	executor *skills.Executor
}

func NewK8sHandler(root *cobra.Command, executor *skills.Executor) *k8sHandler {
	return &k8sHandler{rootCmd: root, executor: executor}
}

func (h *k8sHandler) Register() {
	k8sCmd := &cobra.Command{
		Use:   "k8s",
		Short: "Interact with Kubernetes clusters multi-cloud",
	}

	clustersCmd := &cobra.Command{
		Use:   "clusters",
		Short: "List managed clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			provider, _ := cmd.Flags().GetString("provider")
			res, err := h.executor.Execute(context.Background(), "k8s.list_clusters", map[string]any{"provider": provider})
			if err != nil {
				return err
			}

			payload := res.(map[string]any)
			clusters := payload["clusters"].([]map[string]any)

			for _, c := range clusters {
				fmt.Printf("[%s] Cluster: %s (v%s) Nodes: %d\n", c["region"], c["name"], c["version"], c["node_count"])
			}
			return nil
		},
	}

	k8sCmd.AddCommand(clustersCmd)
	h.rootCmd.AddCommand(k8sCmd)
}
