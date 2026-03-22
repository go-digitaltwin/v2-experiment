package coverage

import "net/netip"

// Endpoint exercises fields with nested stdlib types (multi-segment import
// path like "net/netip").
type Endpoint struct {
	ID     string
	Addr   netip.Addr
	Subnet netip.Prefix
}

// EndpointContainers exercises container fields whose element or key types
// are nested stdlib types.
type EndpointContainers struct {
	ID              string
	Allowlist       []netip.Addr
	PrefixesByRoute map[string]netip.Prefix
	RoutesByPrefix  map[netip.Prefix]string
	Resolvers       [2]netip.Addr
}
