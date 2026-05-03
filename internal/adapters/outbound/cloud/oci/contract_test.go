// File: internal/adapters/outbound/cloud/oci/contract_test.go
// Purpose: Compile-time + runtime guarantees that the OCI adapter satisfies the v1 outbound port contracts.

package oci

import (
	"multix/internal/ports/outbound"
)

var (
	_ outbound.AuthProvider      = (*adapter)(nil)
	_ outbound.InventoryProvider = (*adapter)(nil)
	_ outbound.K8sProvider       = (*adapter)(nil)
)
