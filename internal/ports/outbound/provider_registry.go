package outbound

type ProviderRegistry interface {
	GetAuth(name string) (AuthProvider, error)
	GetInventory(name string) (InventoryProvider, error)
	GetK8s(name string) (K8sProvider, error)
	GetAI(name string) (AIProvider, error)
}
