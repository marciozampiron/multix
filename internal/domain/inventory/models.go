package inventory

import (
	"time"

	"github.com/google/uuid"
)

// Resource represents a discovered infrastructure asset.
type Resource struct {
	ID        string
	AccountID string
	Region    string
	Type      string
	Name      string
	Status    string
	Tags      map[string]string
	CreatedAt time.Time
}

// Summary provides an aggregation of resources by type.
type Summary struct {
	ProviderName string
	Total        int
	CountByType  map[string]int
}

// NewResource initializes a new inventory resource.
func NewResource(accountID, region, resourceType, name string) *Resource {
	return &Resource{
		ID:        uuid.New().String(),
		AccountID: accountID,
		Region:    region,
		Type:      resourceType,
		Name:      name,
		Status:    "UNKNOWN",
		Tags:      make(map[string]string),
		CreatedAt: time.Now(),
	}
}
