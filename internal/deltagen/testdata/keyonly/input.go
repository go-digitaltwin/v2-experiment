// Package keyonly is a test fixture for deltagen. It exercises generation for
// structs with no properties (key fields only); type diversity is covered by
// the coverage fixture.
package keyonly

type Label struct {
	Name string
}
