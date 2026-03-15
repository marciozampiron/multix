package outbound

import (
	"context"

	"multix/internal/domain/inventory"
)

// InventoryProvider defines the required contract for resource querying.
type InventoryProvider interface {
	// ID returns the unique identifier for this provider.
	ID() string

	// List queries and returns a slice of generic resources.
	List(ctx context.Context, resourceType string) ([]*inventory.Resource, error)

	// Scan performs a deep discovery returning a summary of all asset types.
	Scan(ctx context.Context) (*inventory.Summary, error)
}
