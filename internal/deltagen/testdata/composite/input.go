// Package composite is a test fixture for deltagen. It exercises composite-key
// generation; type diversity is covered by the coverage fixture.
package composite

import "net/netip"

type Connection struct {
	TenantID string
	Name     string
	Address  netip.Addr
	Tags     []string
}
