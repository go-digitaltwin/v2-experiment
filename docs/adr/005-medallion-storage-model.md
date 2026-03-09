---
status: accepted
date: 2026-03-03
---

# Adopt the Medallion Storage Model

## Context and Problem Statement

The framework produces two categories of Parquet output: high-frequency
timestamped snapshots and low-frequency authoritative baselines. These serve
different consumers (time-series analytics vs. current-state queries and
recovery) and have different retention characteristics. A storage architecture
is needed to organize them.

## Decision Drivers

- Analytics tools (DuckDB, Spark) work efficiently with Hive-partitioned
  directories.
- Separating raw captures from curated state simplifies both query patterns and
  retention policies.
- The architecture should be familiar to data engineers.

## Considered Options

- **Medallion architecture (bronze/silver)** — two tiers: bronze for timestamped
  captures, silver for latest-state baselines. Each tier uses Hive-style
  partitioning.
- **Single flat directory** — all Parquet files in one directory tree,
  distinguished by naming convention.
- **Database** — persist state in a relational or key-value database instead of
  files.

## Decision Outcome

Chosen option: "Medallion architecture (bronze/silver)."

The two-tier layout maps directly to the framework's two output modes. Bronze
files are immutable append-only captures; silver files are periodically rebuilt
baselines. Hive partitioning by `kind` (entity type) and `date` (calendar day)
enables automatic partition pruning in query engines. The exact path
composition, directory contents (multiple chunks per directory are expected),
and intra-day time partitioning remain to be clarified.

A gold tier is not needed at this stage; downstream consumers apply their own
transformations to bronze and silver data.

### Consequences

- Bronze tier: one Parquet file per entity type per time window, partitioned by
  `kind` and `date`.
- Silver tier: one baseline Parquet file per entity type, rebuilt on each bronze
  commit, partitioned by `kind` and `date`.
- Retention policies are tier-specific and deferred to implementation.

## More Information

- [Storage section in design.md](../design.md#storage)
- [Medallion architecture][medallion] (Databricks glossary)

[medallion]: https://www.databricks.com/glossary/medallion-architecture
