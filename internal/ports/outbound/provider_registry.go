package outbound

type ProviderRegistry interface {
	GetCloudAuthProvider(name string) (AuthProvider, error)
	GetCloudInventoryProvider(name string) (InventoryProvider, error)
	GetKubernetesProvider(name string) (K8sProvider, error)
	GetAIProvider(name string) (AIProvider, error)
}
