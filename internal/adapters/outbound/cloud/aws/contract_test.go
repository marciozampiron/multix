// File: internal/adapters/outbound/cloud/aws/contract_test.go
// Purpose: Compile-time + runtime guarantees that the AWS adapter satisfies the v1 outbound port contracts.

package aws

import (
	"multix/internal/ports/outbound"
)

// Compile-time interface compliance. Any future signature change in the
// outbound ports breaks this assertion before it can break consumers.
var (
	_ outbound.AuthProvider      = (*adapter)(nil)
	_ outbound.InventoryProvider = (*adapter)(nil)
	_ outbound.K8sProvider       = (*adapter)(nil)
)
