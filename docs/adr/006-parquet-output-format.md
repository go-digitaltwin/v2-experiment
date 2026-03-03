---
status: accepted
date: 2026-03-03
---

# Use Parquet as the Output Format

## Context and Problem Statement

The framework persists entity snapshots and baselines to disk. The output format
must support columnar analytics, nested/repeated types, and efficient
compression while remaining accessible to standard query engines.

## Decision Drivers

- Columnar access patterns: analytics queries typically scan a few columns
  across many rows.
- Ecosystem compatibility: DuckDB, Spark, Arrow, and pandas read Parquet
  natively.
- Compression: columnar formats achieve high compression ratios on repetitive
  entity data.
- Nested types: entity models may contain repeated structures (e.g., arrays of
  sub-entities).

## Considered Options

- **Parquet** — columnar, compressed, supports nested/repeated types via the
  Dremel encoding. Wide ecosystem support.
- **CSV/JSON** — human-readable, universally supported, but no columnar access,
  poor compression, and no native nested types.
- **Avro** — row-oriented with schema evolution; better for streaming than
  analytics.
- **Arrow IPC** — columnar and zero-copy, but optimized for in-memory
  interchange rather than long-term storage.
- **Arrow Flight** — a transport protocol built on gRPC and Arrow IPC for
  streaming columnar data between processes. Not a storage format; suited to
  real-time data movement rather than persistence.
- **Vortex** — a next-generation columnar format from SpiralDB (now under LFAI).
  Promises faster scans than Parquet, but the ecosystem is nascent and tooling
  support is limited.

## Decision Outcome

Chosen option: "Parquet."

Parquet provides the best combination of columnar efficiency, compression,
nested type support, and ecosystem reach. The Go library `parquet-go/parquet-go`
handles Parquet I/O; `apache/arrow-go` may complement it later for in-memory
processing.

### Consequences

- All persistent output (bronze and silver) is written as Parquet files.
- Bronze Parquet files are the persistence layer for what is otherwise ephemeral
  NATS output: raw, append-only, retained for later analytics. Retention
  policies are TBD.
- Silver Parquet baselines behave like database tables (one row per entity,
  rebuilt periodically). Table formats such as Apache Iceberg could manage
  silver files in the future (schema evolution, time travel, ACID guarantees),
  but that decision is deferred.
- Entity types with nested structures use Parquet's repeated group encoding.
- Query engines (DuckDB in particular) can read framework output directly
  without ETL.
