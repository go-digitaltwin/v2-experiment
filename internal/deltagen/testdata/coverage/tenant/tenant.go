// Package tenant provides a type for testing intra-module imports in the
// coverage fixture. The deltagen generator should discover its import path
// from the Go module, not from stdlib.
package tenant

// ID identifies a tenant.
type ID string
