---
status: accepted
date: 2026-03-03
---

# Assert/Retract/Ignore Delta Semantics

## Context and Problem Statement

Event sources observe a running system through narrow windows. A single event
rarely reveals the full state of any entity; it provides partial information
where some fields are known and others are absent.

The framework needs semantics for expressing these partial, incremental state
changes. The core problem is choosing a vocabulary for per-field overlay updates
that distinguishes three states: a value is known, a value is no longer valid,
and a value remains unchanged.

See [Why Deltas](../design.md#why-deltas) for the extended motivation.

## Decision Drivers

- Event sources must remain simple and stateless; they should not need to know
  the twin's current state to emit an update.
- Updates are sparse at the column level: a single event may touch one property
  and say nothing about the rest.
- The model must distinguish three semantic states per field: "this value is
  known," "this value is no longer valid," and "this value remains unchanged."

## Considered Options

- **Assert/retract/ignore** — three per-field operations: assert a value,
  retract a previous value, or leave the field untouched.
- **Whole-row replacement** — each update provides the complete state of the
  entity, overwriting all fields.
- **Patch/merge semantics** — JSON Merge Patch or similar: present fields
  overwrite, absent fields are unchanged, explicit null retracts.

## Decision Outcome

Chosen option: "Assert/retract/ignore."

The three-operation model makes the producer's intent explicit at the field
level. Unlike whole-row replacement, it does not require the producer to carry
state. Unlike patch/merge, it distinguishes between "field is unchanged"
(ignore) and "field is explicitly cleared" (retract) without relying on null
semantics that vary across serialization formats.

### Consequences

- Every delta is a struct where each property field wraps one of three states:
  assert(value), retract, or ignore.
- The Entities Service applies these operations deterministically to produce a
  full snapshot after each delta.
- Producers only need to know what the current event tells them; they do not
  track prior state.
