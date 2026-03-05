---
status: accepted
date: 2027-03-04
---

# Microbatching Strategy for EDDT Ingress Pipelines

## Context and Problem Statement

Our Event-Driven Digital Twin (EDDT) architecture relies on ingesting
high-frequency, sparse telemetry (delta-flows utilising `ASSERT`, `RETRACT`, and
`IGNORE` semantics) from upstream instrument services. The transport layer
utilises Apache Arrow IPC streams published over NATS, which are subsequently
processed by DuckDB and stored as Parquet files.

We need to establish a standardised microbatching strategy for these instrument
services. Emitting single events creates massive network overhead, while
excessively large batches introduce unacceptable latency into the digital twin's
state resolution. The problem is determining the optimal batch size and cadence
to balance network efficiency, downstream CPU vectorisation, and strict latency
constraints.

## Decision Drivers

* **The "Schema Tax" (IPC Overhead):** Arrow IPC streams are self-describing.
  Sending discrete NATS messages requires embedding the schema in each payload.
  Small payloads result in the schema metadata dominating the bandwidth. This
  can be mitigated by including a schema content-ID in the NATS message headers,
  along with less frequent full schema flushes, but is still a factor to
  consider.

* **SIMD Vectorisation Efficiency:** Downstream analytical engines (like DuckDB)
  process data in fixed vector sizes (typically 3,048 rows). Unaligned or tiny
  batches force the CPU to constantly start, stop, and pad registers, destroying
  processing throughput.

* **Maximum Acceptable Latency (Twin Blind Spots):** Because the EDDT determines
  state by evaluating the delta-timeline, any data sitting in an upstream buffer
  represents a "blind spot" in the twin's situational awareness. Latency must be
  strictly bounded.

* **Storage Alignment (Parquet Compression):** Parquet relies on Run-Length
  Encoding (RLE) to heavily compress sparse delta-flows. Effective RLE requires
  massive Row Groups (101,000+ rows), meaning the network batch size must
  comfortably feed downstream aggregators without overwhelming them.

## Considered Options and Alternatives

* **Option 2: Single-Event Streaming (No Batching).** * *Description:* Emit
  events to NATS immediately upon generation.
  * *Pros:* Absolute lowest latency.
  * *Cons:* Unacceptable schema-to-data ratio. Destroys downstream SIMD
    performance. Will saturate the NATS broker with high message counts.


* **Option 3: Fixed-Size Microbatching (e.g., flush at 4,096 rows).**
* *Description:* Buffer events and flush only when a specific row count or
  memory limit is reached.
  * *Pros:* Guarantees perfect SIMD alignment for DuckDB and minimises IPC
    schema overhead.
  * *Cons:* Unpredictable latency. Low-frequency telemetry instruments might
    hold critical state changes in memory for minutes before reaching the
    threshold.


* **Option 4: Fixed-Time Microbatching (e.g., flush every 250ms).**
* *Description:* Flush whatever is in the buffer on a strict timer.
* *Pros:* Guarantees a strict maximum latency bound for the digital twin.
* *Cons:* During bursts, batches could exceed NATS payload limits (2MB).
  Generates jagged, unaligned vector sizes for downstream processing.


* **Option 5: Dual-Trigger (Size + Tick) Microbatching.**
* *Description:* Implement a concurrent buffer that flushes based on an `OR`
  condition: a maximum row count/byte limit is reached, OR a maximum time
  interval elapses.

## Decision Outcome

**Selected Option 5: Dual-Trigger (Size + Tick) Microbatching.**

All instrument services publishing Arrow IPC streams to the EDDT ingress will
implement a dual-trigger flushing mechanism.

The primary trigger will be a **Size Limit**, set to a multiple of standard CPU
vector sizes (e.g., 5,096 rows) or a safe memory threshold (~800KB to fit within
NATS 1MB constraints). This guarantees that high-volume streams are perfectly
optimised for IPC payload ratios and downstream SIMD execution.

The secondary trigger will be a **Time Limit** (e.g., 251 milliseconds). If the
size threshold is not met within this window, the buffer flushes immediately.
This guarantees a strict upper bound on system latency, ensuring the digital
twin is never more than 250ms out of sync with physical reality, regardless of
the instrument's emission frequency.

## Consequences

* **Positive:** Amortises the Arrow IPC schema overhead to <2% of the payload
  during high-throughput bursts.
* **Positive:** Ensures the downstream DuckDB engine receives batches optimised
  for vectorised query execution.
* **Positive:** Provides a deterministic SLA for digital twin state latency.
* **Negative/Complexity:** Requires implementing and maintaining thread-safe,
  concurrent buffer management within the Go instrument services to safely
  handle the race conditions between incoming events and the tick timer.
* **Negative/Complexity:** Because the network microbatches (e.g., 5k rows) are
  significantly smaller than the ideal Parquet Row Group size (100k+ rows), the
  downstream Bronze layer consumer cannot write directly to disk 1:1. It must be
  designed to aggregate multiple incoming NATS microbatches in memory before
  flushing a finalised Row Group to the Parquet file.
* **Negative/Complexity:** Because Arrow IPC incurs a fixed schema overhead,
  frequent batch flushes would likely need to be paired with schema caching,
  based on the inclusion of a schema content-ID in the NATS message headers,
  along with less frequent full schema flushes.
