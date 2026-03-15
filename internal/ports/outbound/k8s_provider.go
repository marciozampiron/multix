package outbound

import (
	"context"

	"multix/internal/domain/k8s"
)

// K8sProvider defines the contract for interacting with managed Kubernetes.
type K8sProvider interface {
	// ID returns the unique identifier for this provider.
	ID() string

	// ListClusters retrieves all managed clusters in the current context.
	ListClusters(ctx context.Context) ([]*k8s.Cluster, error)

	// SyncContext downloads the kubeconfig credentials for a specific cluster.
	SyncContext(ctx context.Context, clusterName, region string) error
}
