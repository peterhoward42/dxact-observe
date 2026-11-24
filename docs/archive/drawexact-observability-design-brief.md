> Temporary note from Pete. The decisions here have been folded into the main
> planning docs (`high-level-architecture.md`, `developer-workflow.md`,
> `processing-overview-assumptions.md`, and `event-schema.md`). Prefer those
> documents going forward; keep this file only as the historical source.

# DrawExact Event Observability: Design Brief

## Purpose

Build a small, practical observability system to understand how DrawExact is used. The immediate goal is to capture and examine real user activity, not to build a scalable analytics platform.

At present, usage is very low. If activity eventually grows enough to strain this design, replacing it will be a welcome and acceptable outcome.

## Decisions

- **One event per Google Cloud Storage object.** Each observable action generates an independent write to GCS.
- **Compact JSON payload.** Each object contains a single line of JSON, without unnecessary formatting.
- **Preserve event time.** The payload retains the UTC timestamp recorded when the action occurred. This is authoritative for reconstructing activity, rather than the object's upload or creation time.
- **Use a persistent pseudonymous user ID.** DrawExact mints a UUID on first use in a browser/device environment and reuses it on subsequent visits, subject to the persistence of local browser storage. This is not a verified personal identity.
- **Use a simple date-prefixed, unique object name.** See the convention below. The date prefix supports human inspection and optional date-scoped listing; uniqueness prevents collisions.
- **Analyse in memory.** A Google Cloud Function can list and read stored events, parse them into an in-memory model, group them by UUID, and order them by event timestamp for analysis and UI display.

## Recommended object naming convention

Use the following form:

```text
events/YYYY-MM-DD/<event-uuid>.json
```

For example:

```text
events/2026-11-22/8e6a4c1d-5f29-4e83-9e10-42bd735af0c2.json
```

- **`events/`** keeps observability objects together under one GCS prefix.
- **`YYYY-MM-DD/`** is the UTC calendar date of the event timestamp carried in the JSON payload. This makes manual inspection easier and allows listing a day's events by prefix if useful later.
- **`<event-uuid>`** is a fresh UUID for the event, *not* the persistent pseudonymous user UUID. It makes the object name unique even when multiple events share a timestamp or occur concurrently.
- **`.json`** identifies the payload format.

The original, full-resolution UTC event timestamp remains inside the JSON and is authoritative for analysis. The date in the object name is merely a convenient organizational hint, not a source of event ordering. No hour/minute/second directories, user-based partitions, or other elaborate hierarchy are warranted now.

## Rationale

This approach keeps ingestion simple: no batching, append coordination, message queue, or separate database is required. Independent object writes avoid contention between events.

For the current expected volumes, the operational simplicity matters more than optimizing GCS request charges or the overhead of reading many small objects. Object names are not an essential analytical index because the required event information lives in the payload.

An event timestamp is distinct from storage time. Network delays and retries can cause events to arrive out of order. If unambiguous ordering of events with identical timestamps becomes necessary, a per-user sequence number can be added later.

## Explicitly deferred

- Periodic consolidation of event objects into larger time-window files.
- Export or migration to BigQuery or another analytical store.
- Optimizations for high ingestion rates, large object counts, or selective time-range retrieval.
- A more elaborate naming and partitioning scheme beyond the simple UTC date prefix.

## Guiding principle

**Optimize for understanding DrawExact's first users, not for hypothetical scale.** Keep the system easy to implement, inspect, and replace.
