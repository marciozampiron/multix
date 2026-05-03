// File: internal/adapters/outbound/cloud/gcp/contract_test.go
// Purpose: Compile-time + runtime guarantees that the GCP adapter satisfies the v1 outbound port contracts.

package gcp

import (
	"multix/internal/ports/outbound"
)

var (
	_ outbound.AuthProvider      = (*Adapter)(nil)
	_ outbound.InventoryProvider = (*Adapter)(nil)
	_ outbound.K8sProvider       = (*Adapter)(nil)
)
