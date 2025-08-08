# ADR-001: SQLite Sharding Strategy for Email Archival

## Status

Accepted

## Context

The application is an archival system for Gmail and potentially other messaging services, storing data locally using SQLite. Given long-term growth, the SQLite database will be **sharded** across multiple `.db` files to:

* Manage file size per shard (ensuring performance and stability)
* Enable modular storage with partial backups or sync
* Avoid SQLite limitations on table/index sizes and write contention

Several design tradeoffs were considered for **how to handle querying across shards**, especially for **tag data**, which may change over time.

## Decision

We will:

1. Use **multiple SQLite databases (shards)**, each with an identical schema, where each shard stores a subset of the emails.
2. Use SQLite’s built-in `ATTACH DATABASE` functionality to query across shards within a single connection.
3. **Re-attach all shards on application startup**, since `ATTACH` is scoped to a single connection and is not persistent across sessions.
4. Avoid connection pooling or ensure each connection is pre-warmed with the correct `ATTACH` commands.
5. Keep **tagging data isolated to each shard**, rather than normalizing globally across shards. This ensures:

    * Tag associations reflect historical context and do not change retroactively.
    * Each shard remains logically self-contained and auditable.
6. Avoid centralized `tags` or `tag_id` tables that span shards, to preserve historical tagging integrity.
7. Perform **cross-shard queries using `UNION ALL`**, and merge results in application code.
8. Optionally maintain a `meta.db` file to record shard file paths and other metadata, for dynamic attachment.

## Rejected Alternatives

* **Duplicating the `tags` table across shards** with identical `tag_id` values:

    * Would require enforcing consistency of tag names and IDs across all shards.
    * Fails to capture the reality that tag definitions and usage may evolve over time.
* **Centralized `tags` database (shared foreign key model):**

    * Introduces complexity in managing foreign key relationships across files.
    * Violates the goal of immutability and self-containment for shards.
* **Storing tags as delimited strings in a `TEXT` column:**

    * Conflicts with the normalized schema design.
    * Inhibits efficient querying and increases storage redundancy.

## Consequences

* Application logic must explicitly handle:

    * Attaching databases on startup
    * Running `UNION ALL` queries across shards
    * Aggregating and filtering results in memory
* Enables independent tagging semantics and historical accuracy per shard.
* Maintains simplicity and SQLite compatibility, with no external query federation layer required.
* Enables long-term scalability to dozens of shards while retaining strong control over performance and storage efficiency.

## Future Considerations

* If the number of shards grows significantly (>100), we may need to:

    * Use dynamic attachment (attach only relevant shards at query time)
    * Create summary/index tables in a central DB for faster lookups
* Consider adding sharding heuristics (by year, sender domain, size thresholds, etc.)
