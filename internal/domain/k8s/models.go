package k8s

// Cluster represents a Kubernetes cluster discovered via a cloud provider.
type Cluster struct {
	ID        string
	Name      string
	Region    string
	Status    string
	Version   string
	NodeCount int
}

// Namespace represents a namespace inside an active cluster.
type Namespace struct {
	Name   string
	Status string
}
