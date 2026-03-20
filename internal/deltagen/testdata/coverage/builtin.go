// Package coverage is a test fixture for deltagen. It exercises type diversity
// across multiple files and types, verifying that the generator handles
// containers (maps, slices, arrays), external imports, and same-package types.
//
// Each file adds types for a specific import kind. The structural fixtures
// (simple, composite, keyonly) cover key shapes and the apply flag.
package coverage

// Metric exercises container field types with builtin element types only.
// No imports are needed in the generated output.
type Metric struct {
	ID      string
	Values  map[string]float64
	Buckets [8]float64
	Labels  []string
	Tags    map[string]string
}
