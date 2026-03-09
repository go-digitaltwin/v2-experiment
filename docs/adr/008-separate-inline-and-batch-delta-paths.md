---
status: accepted
date: 2026-03-04
---

# Separate Inline and Batch Processing Paths for Deltas

## Context and Problem Statement

The framework receives deltas (partial entity updates) that serve two distinct
purposes:

1. **Inline processing**: apply deltas to in-memory state and emit full
   snapshots immediately, providing a live view of the world.
2. **Batch persistence**: accumulate deltas and flush them to Parquet files in
   the bronze tier, providing a durable analytical record.

These two purposes have fundamentally different operational characteristics.
Inline processing is stateful, memory-bound, and latency-sensitive. Batch
persistence is I/O-bound, throughput-oriented, and driven by window boundaries
rather than individual messages. The question is whether these concerns should
be coupled or independent.

## Decision Drivers

- **Different resource profiles**: inline processing holds volatile state per
  entity in RAM; batch persistence buffers messages and performs filesystem
  writes. Memory consumption in the inline path is proportional to entity count;
  in the batch path it is proportional to throughput times window size.
- **Different scaling models**: inline processing can be partitioned by entity
  kind across separate processes, giving convenient control and observation of
  memory consumption per kind. Batch persistence is constrained to a single
  writer per file, though different kinds land in different files (Hive
  partitioning), so per-kind parallelism is natural.
- **Different ordering constraints**: inline processing must apply deltas in
  per-entity causal order to maintain correct state. Batch persistence stores
  timestamped rows; within a Parquet file rows may be unordered, sortable at
  query time.
- **Different rates of advancement**: inline processing emits a snapshot for
  every delta received. Parquet files advance on window boundaries (time or size
  thresholds). The two cadences are independent.
- **Failure isolation**: a failure in one path should not disrupt the other.

## Considered Options

- **Two independent paths** — deltas fan out to inline processing and batch
  persistence as independent consumers. Neither path depends on the other for
  delta handling.
- **Sequential pipeline** — all deltas pass through inline processing first;
  batch persistence operates only on the inline path's output (snapshots). Raw
  deltas are not persisted independently.
- **Single combined service** — one process handles both in-memory state
  management and Parquet persistence, sharing memory and a crash domain.

## Decision Outcome

Chosen option: "Two independent paths."

The delta input fans out to two independent processing paths. Each path consumes
the full delta stream and operates at its own pace, with its own failure domain,
and under its own scaling constraints. The two paths share nothing beyond the
delta input.

This independence means that the framework requires a message distribution
mechanism to deliver deltas to both paths (a message broker, gRPC fan-out, or
similar). The choice of mechanism is a separate decision.

### Consequences

- The inline path applies deltas to in-memory entity state and emits snapshots.
  It is stateful, latency-sensitive, and partitionable by entity kind as
  separate processes for memory isolation.
- The batch path accumulates deltas in windows and flushes them to bronze
  Parquet files. It is I/O-bound and advances on its own schedule, independent
  of the inline path's output rate.
- Parquet files and live snapshots advance at different rates. There is no
  guarantee that the latest bronze file reflects the same point in time as the
  latest live snapshot.
- Ordering diverges by design: the inline path enforces per-entity causal order;
  the batch path stores timestamped rows that may be unordered within a file.
- Failure isolation: an inline-path crash does not prevent delta persistence,
  and a batch-path crash does not disrupt live snapshot output.
- **Asymmetry with snapshots**: snapshot batching (also in bronze, per
  [ADR-007](007-dual-bronze-content.md)) depends on the inline path's output,
  since snapshots originate there. A stalled inline path blocks snapshot
  persistence but not delta persistence. Delta batching is fully independent of
  the inline path.
- A sequential pipeline was rejected because it couples delta persistence to the
  health and throughput of the inline path. A combined service was rejected
  because it merges two fundamentally different resource profiles (memory-bound
  vs. I/O-bound) into one crash domain.

## More Information

- [Architecture in design.md](../design.md#architecture)
- [ADR-007: Store Both Deltas and Snapshots in the Bronze Tier](007-dual-bronze-content.md)
