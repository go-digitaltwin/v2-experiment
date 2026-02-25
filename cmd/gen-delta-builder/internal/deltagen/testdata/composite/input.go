package composite

import "net/netip"

type Connection struct {
	TenantID string
	Name     string
	Address  netip.Addr
	Tags     []string
}
