// gen-delta-builder reads a Go struct type and generates a companion
// *TDelta builder with per-field Set, Clear, and Skip methods.
//
// # Usage
//
//	gen-delta-builder -type T -key F1[,F2,...] [-apply] [-input dir] [-output path]
//
// The generated code targets pkgsite readability: every exported symbol carries
// a doc comment with cross-references back to the source type. Method bodies
// are stubs; the internal representation is provided by a separate runtime
// package.
//
// # Flag design
//
// The current flag set reflects a deliberate balance between fixed output and
// caller control. Some parts of the generated code are always emitted (the
// delta struct, constructor, key accessors, and the three per-property methods);
// others are opt-in (-apply). This balance is not accidental: the fixed core
// is what every delta builder needs, while -apply demonstrates the opt-in
// pattern for features that not every entity requires.
//
// The flag surface is designed to grow in a backwards-compatible direction:
//
//   - Required flags can become optional. -key is required today because every
//     delta builder needs at least one key field. If a future convention
//     (struct tags, marker interfaces) can infer keys, -key becomes optional
//     without breaking existing go:generate directives.
//
//   - New opt-in flags follow the -apply pattern. Features like getter
//     generation or tombstone methods can be added behind new boolean flags
//     that default to false; existing invocations are unaffected.
//
//   - Opt-out flags can exclude individual fields. A future -skip flag (or
//     struct-tag annotation) could suppress generation for specific properties,
//     letting callers narrow the generated API without touching the source
//     struct.
//
// Each of these extensions is additive: no existing directive needs to change
// when a new flag appears.
package main
