// File: internal/domain/network/models.go
// Company: Hassan
// Creator: Zamp
// Created: 03/05/2026
// Updated: 03/05/2026
// Purpose: Domain types for AI-generated network topologies (infra.generate_network).

// Package network defines provider-agnostic network topology types produced
// by the infra.generate_network skill. The shape mirrors a 3-tier VPC pattern
// (public / private / database) so it maps cleanly onto AWS, GCP and OCI.
package network

// Subnet describes a single network slice within a VPC.
type Subnet struct {
	Name string `json:"name"`
	CIDR string `json:"cidr"`
	AZ   string `json:"availability_zone"`
	Tier string `json:"tier"` // "public" | "private" | "database"
}

// RouteRule describes a directional route between two named subnets/gateways.
type RouteRule struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Action string `json:"action"` // "allow" | "deny" | "nat"
}

// NetworkSpec is the top-level shape returned by infra.generate_network.
// It is intentionally provider-agnostic: rendering into Terraform / CloudFormation
// / OCI Resource Manager is a downstream concern.
type NetworkSpec struct {
	ProviderName string      `json:"provider"`
	Region       string      `json:"region"`
	VPCCidr      string      `json:"vpc_cidr"`
	Subnets      []Subnet    `json:"subnets"`
	RouteRules   []RouteRule `json:"route_rules"`
}
