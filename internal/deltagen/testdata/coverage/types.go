package coverage

import "syscall"

// Status is a sibling type (defined in another file within the same package)
// with builtin underlying type.
type Status int

// Token is a sibling type (defined in another file within the same package)
// with empty struct underlying type.
type Token struct{}

// Signal is a sibling type (defined in another file within the same package)
// whose definition references a stdlib type (syscall.Signal). Using it as a
// field should not emit an import for "syscall" in the generated code; the
// generator sees it as a local type.
type Signal syscall.Signal
