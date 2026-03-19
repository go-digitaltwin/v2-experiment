// Package simple is a test fixture for deltagen. It exercises single-key
// generation with builtin types only; type diversity is covered by the
// coverage fixture.
package simple

type Device struct {
	ID    string
	Name  string
	Model string
}
