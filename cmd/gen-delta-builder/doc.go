// gen-delta-builder reads a Go struct type and generates a companion
// *TDelta builder with per-field Set, Clear, and Skip methods.
//
// # Usage
//
//	gen-delta-builder -type T -key F1[,F2,...] [-apply] [-dir path] [-output path]
//
// The generated code targets pkgsite readability: every exported symbol carries
// a doc comment with cross-references back to the source type. Method bodies
// are stubs; the internal representation is provided by a separate runtime
// package.
//
// # Compatibility
//
// The generator currently emits a fixed set of symbols for every invocation:
// the delta struct, constructor, key accessors, and the three per-property
// methods (Set, Clear, Skip). The -apply flag is the only opt-in; it adds an
// Apply method for entities with straightforward update semantics.
//
// This surface can evolve without breaking existing go:generate directives
// because every planned change is additive:
//
//   - New opt-in flags follow the -apply pattern. Features like getter
//     generation or tombstone methods appear behind boolean flags that default
//     to false. Existing directives produce the same output as before.
//
//   - Opt-out flags suppress parts of the fixed output using the -no- prefix
//     convention (as in git --no-verify, --no-edit). For example, -no-clear
//     could suppress Clear method generation for callers that don't need
//     retraction. The default remains to generate everything.
//
//   - Required flags can become optional. -key is required today because every
//     entity has an identifying key, but if a future convention (struct tags,
//     marker interfaces) can infer keys, the flag becomes optional without
//     invalidating directives that still pass it explicitly.
//
//   - Struct-tag annotations can narrow generation per field. A tag like
//     `delta:"-"` could exclude a specific property from the generated API
//     without requiring flag changes at the call site.
//
// In all cases, a directive that works today continues to produce identical
// output after the generator is extended.
package main
