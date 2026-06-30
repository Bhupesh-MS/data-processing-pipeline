# Data Processing Pipeline - Design Report

## Architecture Overview
The system is built as a highly concurrent data processing pipeline leveraging Go's native goroutines and channels to implement a multi-stage fan-out/fan-in architecture. The overarching goal is to safely process potentially massive streams of data from various inputs (CSV, JSON, APIs) through validation, transformation, aggregation, and export stages without holding all records in memory.

## Concurrency Model
The pipeline uses **channels** to move typed `Record` structures between stages:
- `recordsCh` -> `validatedCh` -> `transformedCh` -> `resultCh` (or aggregation then `resultCh`) -> Export.

1. **Ingestion (Fan-out)**: Multiple source configurations spawn independent goroutines. As they read data (e.g. streaming a remote CSV or parsing JSON tokens), they push standardized `Record`s onto `recordsCh`.
2. **Validation & Transformation (Worker Pools)**: A configurable number of goroutines (default 5) listen on `recordsCh` and `validatedCh` respectively. This allows CPU-bound validation and transformation rules to operate in parallel, significantly boosting throughput.
3. **Aggregation (Fan-in)**: A single aggregator goroutine collects records from `transformedCh`, building an in-memory map of sums, counts, and averages grouped by specified fields. Upon channel closure (completion of upstream processing), it flushes aggregated records to `resultCh`.
4. **Export**: The consumer of `resultCh` orchestrates writing the results simultaneously to SQLite tables, CSV, or JSON file descriptors.

## Trade-offs & Decisions
1. **Dynamic Typing**: `Record.Attributes` is defined as `map[string]any`. While this sacrifices some compile-time type safety, it is essential for a generic pipeline that doesn't know schema definitions until runtime.
2. **In-Memory Aggregation**: The aggregation state maps are stored in-memory. For extremely high-cardinality datasets (millions of unique group keys), this could cause memory pressure. A persistent KV store like Redis could be substituted for the `state` maps if necessary, trading network latency for unbounded memory scalability.
3. **SQLite**: Chosen for local job metadata and error persistence due to its simplicity. A production environment with multiple instances of this microservice would require Postgres, but the codebase uses standard `database/sql` making migration trivial.
4. **Error Collection**: Instead of halting the pipeline on a bad record, errors are pushed to a non-blocking `errCh`. A dedicated collector goroutine writes these safely to the DB, ensuring resilience.
