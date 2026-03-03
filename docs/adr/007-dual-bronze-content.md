---
status: accepted
date: 2026-03-03
---

# Store Both Deltas and Snapshots in the Bronze Tier

## Context and Problem Statement

The bronze tier captures entity state within time windows. Two data streams flow
through the framework's core: raw deltas (the input to the Entities Service) and
full-state snapshots (the output). The choice of which to persist in bronze
affects completeness, query complexity, and downstream processing.

## Decision Drivers

- Completeness: persisting both the ephemeral input and the materialized output
  gives analytics consumers the full picture.
- Query simplicity: analysts should be able to query bronze snapshots directly
  without reconstructing state from a chain of deltas.
- No replay required: pre-computed snapshots free consumers from delta-replay
  logic.
- Storage cost: full snapshots are larger than sparse deltas, but columnar
  compression mitigates the difference.

## Considered Options

- **Snapshots only** — every bronze row contains the complete state of one
  entity at one point in time. Consumers query directly; diffs computed on the
  fly.
- **Deltas only** — bronze rows contain only the raw deltas as received from the
  event source. Consumers must replay deltas to reconstruct state.
- **Both** — write both delta batches and snapshot batches per window,
  preserving the raw input alongside the materialized output.

## Decision Outcome

Chosen option: "Both."

The bronze tier stores both raw delta batches and full-state snapshot batches.
Both are "raw" in the medallion sense: they are direct captures of ephemeral
NATS traffic, retained as the persistence layer for what would otherwise be
lost. Snapshots are more queryable than deltas (columnar, full rows per entity),
but are not silver because a single batch may contain multiple rows for the same
entity across the window.

### Consequences

- Each bronze window may contain a delta batch and a snapshot batch, one file of
  each type per entity type.
- Snapshot batches contain one row per entity per observation, with all
  properties populated. Columnar compression mitigates the storage overhead of
  repeated unchanged values.
- Delta batches preserve the raw input for auditing, replay, and fine-grained
  change analysis.
- Downstream consumers choose their access pattern: query snapshots for
  convenience, or deltas for precision.
- Silver baselines are built from bronze snapshots (latest snapshot per entity
  wins), not by replaying deltas.

## More Information

- [Bronze Tier in design.md](../design.md#bronze-tier)
