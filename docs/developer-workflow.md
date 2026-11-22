# Developer workflow

- This project is produced with phased scoping, design, planning and
  implementation across multiple sessions.
- The convergence of design and planning docs will require manual iteration.

## Automation considered

- We considered driving the first full implementation with a carefully described
  phased automation script that would govern AI work over multiple sessions
  (phases sized for one session each; a scripted run on the order of a couple
  of hours).
- We are not using that approach for now. Work proceeds manually in incremental
  phases with tests and reviews between them.

## Already decided

- Design events before building the service
- Prove one Go endpoint can take POSTs and serve a Svelte UI
- Keep business logic off HTTP and cloud storage details
- Analyse via an in-memory model, not by querying raw storage directly
- `dxact-draw` owns what gets emitted
- Work manually in phases, not via a big automation script

## Still open

- Hourly JSONL append versus one object per event (see high-level-architecture.md)
- Exactly how the UI loads stored events into the in-memory model
- Which abuse protections to turn on before public use (see access-control.md)

## Phases

### 1. Event design

Agree the product questions we care about, then the event names and fields that
would answer them. The questions are in
[product-questions.md](./product-questions.md). Keep names plain. Prefer a few
generic events with parameters (for example "command executed" plus the command
name) over dozens of one-off types. Agree the common envelope: type, UTC time,
browser UUID, schema version, and extra JSON fields. When DrawExact emits, that
contract lives in `dxact-draw`.

**Done when:** there is a written list of questions, events, and fields we are
willing to implement. The questions are `product-questions.md`; the draft
events and fields are `event-catalogue.md` (not final until we agree to
implement that list).

### 2. Local ingest slice

Build the first real code entirely on a developer machine: a Go handler that
accepts a POST and serves a Svelte "Hello World" on GET, real validation of the
event envelope (including schema version), and saves through a storage interface
into an in-memory fake. The HTTP layer stays thin; the save logic does not know
about GCS. UUIDs are not created here — only checked on the payload. No Google
Cloud in this phase — that waits until Phase 3. This proves the shape (one
service, both jobs) and that ingestion works under test, not analytics.

**Done when:** POST and GET work locally with the trivial UI; valid events are
accepted and stored in the fake; bad ones are rejected; tests cover that without
cloud.

### 3. Cloud storage

First contact with Google Cloud: stand up the project enough to deploy the
Cloud Function and use GCS. Investigate and choose: append lines into a
time-window object (with the concurrent-overwrite risk), or one object per event
under a time-based prefix. Implement the GCS backend behind the same interface
used by the Phase 2 fake. Object names stay time-based; do not put event type or
user id in the path.

**Done when:** the choice is written down, the function deploys, GCS save works,
and tests (plus a small real check) show events survive.

### 4. DrawExact sends events

In the browser: create a UUID on first visit, keep it in localStorage, send it
on every event. Point `dxact-draw` / `dxact-wasm` at the new service using the
Phase 1 schema.

**Done when:** a normal use of DrawExact produces events visible in storage.

### 5. In-memory analysis model

Load stored events into a model built for questions, not for cheap storage —
indexes, maps, or similar as needed. Support older schema versions on the way
in; ignore obsolete fields; do not rewrite old objects. Test queries with
fixture events so this phase does not depend on live traffic. May overlap with
Phase 4 once Phases 2–3 are solid.

**Done when:** at least a couple of real questions are answered correctly in
tests from known inputs.

### 6. First observability UI

Replace Hello World with screens for those first hard-coded recipes. Load the
model when the page opens; refresh means reload. Charts and richer filters wait
until we know which views we actually use.

**Done when:** we can open the UI and read answers to the first questions from
real or fixture-backed data.

### 7. Production hardening

Not "start testing" — earlier phases should already leave the code trustworthy.
Add whatever is still needed before a public ingest URL: rate limits, request
size limits, cloud quotas, a proper production deploy, and a thin end-to-end
check. Phase 3 already puts a minimal version on Google Cloud with GCS; this
phase is the real service.

**Done when:** we are willing to point real DrawExact traffic at it and use the
UI as the daily read.
