# High Level Architecture

## Purpose

Build a small, practical observability system to understand how DrawExact is
used. The immediate goal is to capture and examine real user activity, not to
build a scalable analytics platform.

At present, usage is very low. If activity eventually grows enough to strain
this design, replacing it will be a welcome and acceptable outcome.

**Guiding principle:** optimize for understanding DrawExact's first users, not
for hypothetical scale. Keep the system easy to implement, inspect, and
replace.

## One Cloud Function, one origin

Build a small, self-contained Google Cloud Function that both receives events
and serves a Svelte single-page application (SPA). Keep the Go Functions
Framework and source-deployment workflow, without introducing a separately
managed container image or frontend hosting service.

A single HTTP Cloud Function exposes one origin:

- **POST `/`** receives events, validates and processes them, and writes to
  GCS (see storing telemetry events below).
- **GET `/`** and **HEAD `/`** serve the Svelte application's `index.html`.
- **GET `/assets/...`** and **HEAD `/assets/...`** (and other real static
  asset paths) serve the compiled JavaScript, CSS, and other assets.
- **GET on SPA navigation paths**, if client-side routing is used, falls back
  to `index.html` when the path is not a real asset — without hiding missing
  asset errors (missing assets still return actual 404s).
- Unsupported methods return `405 Method Not Allowed` with an appropriate
  `Allow` header.

The browser runs the Svelte application; Go only serves its compiled assets and
implements the server-side endpoints. Requests from the UI to the backend can
use the same origin, avoiding cross-origin configuration for that interaction.

Hard-coded analysis recipes are shown by the SPA (client-side routes and
layouts). Any later same-origin GET APIs for query results are separate from
serving HTML. Use `embed.FS`, `fs.Sub`, and `http.FileServer` as the starting
point for serving the compiled frontend, with explicit SPA fallback behaviour
if needed.

### Why this shape

- **Minimal infrastructure:** one deployed function handles the UI and event
  ingestion.
- **Familiar operational model:** preserves Go, the Functions Framework,
  Linux-like development environments, Make automation, and Google Cloud
  source deployment.
- **One coherent release:** frontend and backend are built and deployed
  together from a single repository revision.
- **Simple browser integration:** UI requests can reach the backend at the
  same origin.
- **Small-project fit:** avoids the coordination overhead of multiple
  repositories and independent hosting/deployment pipelines.
- **Clear ownership:** Svelte owns frontend source and compilation; Go/Make
  own assembly, verification, and deployment.

Layout and the working decision to embed the SPA are in
technology-choices.md. Make-driven build/deploy is in developer-tooling.md.

### Local proof, then cloud

Phase 2 implements this decided shape locally (POST plus serving the UI from
one process) under test. The first Google Cloud deploy waits until Phase 3
with GCS. That local work is a proof the arrangement runs, not a choice of
architecture.

### Trade-offs and implementation checks

- **Cold starts:** a function scaled to zero may delay the initial page load;
  minimum instances are an option if this matters.
- **Static serving cost:** each asset request invokes the function. For
  substantially higher traffic, separate static hosting or a CDN may become
  preferable.
- **Caching:** hashed Vite assets can receive long-lived cache headers;
  `index.html` should be refreshed more conservatively.
- **Security:** if public GET access requires an unauthenticated function,
  POST ingestion needs its own authentication or request verification as
  appropriate. HTTP method separation alone is not a security boundary — see
  access-control.md.
- **Deployment coupling:** a UI-only change still requires redeploying the
  function; for this small project, the simplicity of one release is likely
  worth it.
- **Routing:** static asset requests should return actual 404s when missing,
  while intentional SPA navigation routes can fall back to `index.html`.
- **Reproducibility:** commit the npm lockfile, use a consistent Node
  version, and ensure Make stages the assets before deployment (see
  developer-tooling.md).

## HTTP handling and application logic

- Register one HTTP handler with the Go Functions Framework. Dispatch by HTTP
  method and path as above.
- The HTTP handler is a thin entry point: it receives the request, interprets
  it, and passes the relevant information to ordinary Go application code.
- Application logic must not depend directly on HTTP handling or on Google
  Cloud Storage.
- Storage is behind an interface. Two implementations:
	- in-memory / fake — for tests
	- Google Cloud Storage — for production
- The application logic operates identically with either implementation, so
  almost all of the backend can be tested without deploying to Google Cloud.

## POST ingestion

For each POST, the handler:

1. Receives the event
2. Validates its contents
3. Converts it into the appropriate storage representation
4. Writes it to storage
5. Returns success or failure

## Storing telemetry events

- Incoming events are stored in GCS.
- **One event per GCS object.** Each observable action generates an
  independent write. There is no batching, append coordination, message
  queue, or separate database.
- **Compact JSON payload.** Each object contains a single line of JSON,
  without unnecessary formatting.
- **Preserve event time.** The payload retains the UTC timestamp recorded
  when the action occurred. That timestamp is authoritative for
  reconstructing activity, rather than the object's upload or creation time.
- The GCS bucket itself is a project setup detail. Observability objects live
  under a shared object-name prefix (see naming below), not by naming the
  bucket `events`.
- Do not encode event type or the persistent user identifier into object
  names.

### Object naming convention

Use the following form:

```text
events/YYYY-MM-DD/<event-uuid>.json
```

For example:

```text
events/2026-11-22/8e6a4c1d-5f29-4e83-9e10-42bd735af0c2.json
```

- **`events/`** keeps observability objects together under one GCS prefix.
- **`YYYY-MM-DD/`** is the UTC calendar date of the event timestamp carried in
  the JSON payload. This makes manual inspection easier and allows listing a
  day's events by prefix if useful later.
- **`<event-uuid>`** is a fresh UUID for the event, *not* the persistent
  pseudonymous user UUID. It makes the object name unique even when multiple
  events share a timestamp or occur concurrently. Minting this id is a
  storage/naming concern on the ingest side; it is distinct from the user
  UUID in the payload.
- **`.json`** identifies the payload format.

The original, full-resolution UTC event timestamp remains inside the JSON and
is authoritative for analysis. The date in the object name is merely a
convenient organizational hint, not a source of event ordering. No
hour/minute/second directories, user-based partitions, or other elaborate
hierarchy are warranted now.

### Rationale

Independent object writes keep ingestion simple and avoid contention between
events (including the concurrent-overwrite risk that shared append blobs
would have created).

For the current expected volumes, operational simplicity matters more than
optimizing GCS request charges or the overhead of reading many small objects.
Object names are not an essential analytical index because the required event
information lives in the payload.

An event timestamp is distinct from storage time. Network delays and retries
can cause events to arrive out of order. If unambiguous ordering of events
with identical timestamps becomes necessary, a per-user sequence number can
be added later.

### Explicitly deferred

- Periodic consolidation of event objects into larger time-window files.
- Export or migration to BigQuery or another analytical store.
- Optimizations for high ingestion rates, large object counts, or selective
  time-range retrieval.
- A more elaborate naming and partitioning scheme beyond the simple UTC date
  prefix.

## Serving the UI

- We will define a set of query shapes over the event store
- This will start as a small set and will change significantly over time
- The results of each such query shape / recipe will be displayed by whatever
  layout serves that query well (in the SPA)
- Analysis loads events into an in-memory model (list and read stored objects,
  parse them, group by user UUID, order by event timestamp) for analysis and
  UI display — see processing-overview-assumptions.md
