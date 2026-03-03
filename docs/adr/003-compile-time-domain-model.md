---
status: accepted
date: 2026-03-03
---

# Define Domain Models at Compile Time

## Context and Problem Statement

The digital twin framework needs to know the shape of each entity type (primary
keys, properties, relationships) in order to apply deltas, emit snapshots, and
write Parquet files. This schema can be provided at compile time (Go structs) or
at runtime (configuration files, reflection).

## Decision Drivers

- Type safety: delta builders and snapshot accessors should be checked by the
  compiler.
- Performance: avoiding reflection on the hot path (delta application) matters
  for throughput.
- Developer experience: Go structs are familiar, IDE-navigable, and
  self-documenting.
- Extensibility: runtime-defined schemas would allow dynamic onboarding of new
  entity types without recompilation.

## Considered Options

- **Compile-time Go structs** — entity types defined as exported Go types; code
  generation produces delta builders and Parquet serializers.
- **Runtime schema definition** — entity types loaded from YAML/JSON
  configuration at startup; all operations use reflection or generic maps.
- **Hybrid** — compile-time structs as the primary path, with a future runtime
  path for dynamic schemas.

## Decision Outcome

Chosen option: "Compile-time Go structs" with an acknowledgement that a runtime
path may be layered on later.

Compile-time definition maximizes type safety and performance. Code generation
bridges the gap between struct definitions and the framework's internal
machinery (delta builders, Parquet writers). The trade-off is that adding a new
entity type requires recompilation, which is acceptable for an experimental
project and aligns with Go's philosophy.

### Consequences

- Entity types are Go structs annotated with struct tags for key and property
  metadata. Metadata may also be provided through explicit type registration
  with the framework (a registration API called in Go code). Both mechanisms are
  compile-time: the compiler checks them, and no runtime schema loading is
  required.
- A code generator produces per-entity delta builders and snapshot types.
- Adding or modifying entity types requires recompilation and redeployment.
