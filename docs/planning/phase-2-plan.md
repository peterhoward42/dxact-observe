# Phase 2 — Local ingest: build the first working telemetry receiver

The overall phase plan is in [developer-workflow.md](../developer-workflow.md). Phase 1 (event design) is complete. This document explains what to build in Phase 2 and how to verify it.

## Goal

Currently, `dxact-observe` contains documentation but no application code. Phase 2 establishes the first working version of the telemetry receiver, entirely locally.

By the end of this phase, one Go HTTP Cloud Function entry point will be able to:

- Receive telemetry events as JSON via HTTP POST requests.
- Check that each event contains the required information in the correct format.
- Save valid events in an in-memory store (a temporary substitute for Google Cloud Storage, or GCS).
- Serve a basic Svelte webpage displaying Hello World via HTTP GET requests.

The aim is to prove that these pieces work together before introducing Google Cloud or connecting DrawExact itself to the receiver. Events stored in memory will not survive restarting the application.

## How the pieces fit together

There are two main paths through the application:

```text
POST / → Application → validate event → Store → HTTP response
GET /  → Application → frontend filesystem → webpage
```

Both requests enter through the same HTTP handler in `function.go`. That file assembles the application and hands requests over to it; it does not contain the request-handling logic itself.

The application uses a storage interface rather than talking directly to GCS. This means a future GCS implementation can replace the in-memory store without the event-receiving logic needing to know how or where events are saved.

## Target directory structure

```text
dxact-observe/
├── Makefile
├── go.mod
├── go.sum
├── function.go                 # Cloud Function entry point
├── internal/
│   ├── app/                    # Application, Dependencies, HTTP request handling
│   ├── event/                  # Event structure and validation
│   └── storage/                # Store interface and in-memory implementation
├── ui/                         # Svelte frontend source, including .nvmrc
├── static/                     # Generated frontend files for Go to embed
└── docs/
```

Each part has a distinct responsibility:

| Location | Responsibility |
| --- | --- |
| `function.go` | Assemble the runtime dependencies, embed `static/`, construct `Application`, register the HTTP handler with the Functions Framework and delegate requests. |
| `internal/app` | Interpret HTTP requests, coordinate event validation and saving, and serve the frontend. Receive the frontend filesystem through `Dependencies`, rather than embedding it directly. |
| `internal/event` | Define and validate the common event envelope described in [event-schema.md](../event-schema.md). |
| `internal/storage` | Define the `Store` interface, implement the in-memory store and construct event object names. |
| `ui/` | Contain the minimal Svelte Hello World application. |
| `static/` | Contain the generated frontend files copied from `ui/dist/` by Make. Nothing in this directory is committed to Git. |

The dependency direction is `function.go` → `internal/app` → `internal/event` / `internal/storage`. Storage must never import app. Keeping these responsibilities separate allows the application to be tested without the Cloud Function framework or a real frontend build.

## Implementation steps

### 1. Set up the Go module and build commands

Create `go.mod`, `go.sum` and a root `Makefile`. Follow [developer-tooling.md](../developer-tooling.md) for the build conventions.

There should be two ways to test the project:

**Quick backend tests:** run `go test ./internal/...`. These tests must not require Node.js, Vite or a built frontend.

**Full verification:** install the frontend dependencies, build the Svelte application, copy the contents of `ui/dist/` into `static/`, and then run Go build/test checks for the whole module, including the root package.

Why the distinction? The root package embeds the built frontend into the Go application. Go cannot compile that package until the frontend files exist. Testing only the internal packages avoids having to build Svelte every time a backend change is made.

The future deploy target will also build and stage the frontend before uploading the source. Actual deployment is not part of Phase 2.

Add these directories to `.gitignore`:

- `static/`
- `ui/node_modules/`
- `ui/dist/`

Do not commit anything under `static/`, not even a placeholder file. Make will generate the directory and its contents when required.

Pin Node.js **22** in `ui/.nvmrc` so frontend builds use a consistent Node version.

### 2. Define and validate telemetry events (`internal/event`)

Every event has a common set of fields, called the **event envelope**. The envelope is already specified in [event-schema.md](../event-schema.md); this step implements that specification in Go.

| Field | Validation requirement |
| --- | --- |
| `time` | RFC 3339 timestamp with `Z` indicating UTC. |
| `type` | Non-empty string. |
| `browser_uuid` | Well-formed UUID supplied by the browser. |
| `schema_version` | Must equal `1`. |
| `params` | Must be a JSON object; an empty object (`{}`) is valid. |

The order of fields in an incoming POST does not matter.

The receiver validates the browser UUID supplied in the payload; it does not generate or replace it.

For Phase 2, validate only the common envelope. Do not yet check whether the parameters for each individual Now event match the requirements in [event-catalogue.md](../event-catalogue.md). In other words, check that `params` is an object, but not whether it contains all the fields expected for a particular event type.

### 3. Create the storage interface and in-memory store (`internal/storage`)

Define a `Store` interface that provides the ability to save one validated event. An interface describes what the storage component must do without tying the application to a particular storage technology.

Implement an in-memory version of `Store`. It will be used both by tests and by the locally running application in this phase. Do not introduce a GCS client.

#### Stored data

Save each validated envelope, including `params`, as compact, single-line JSON, with `time` as the first field. Save the validated data rather than the original raw HTTP request body.

#### Object names

Give each stored event an object name in this format:

```text
events/YYYY-MM-DD/<event-uuid>.json
```

For example:

```text
events/2026-11-24/550e8400-e29b-41d4-a716-446655440000.json
```

The date comes from the UTC `time` in the event payload, not from the time the server receives it.

Generate a fresh event UUID for each object name. This is distinct from `browser_uuid`: the browser UUID identifies a browser, whereas the event UUID distinguishes individual stored events. Multiple events from the same browser must not overwrite one another.

The logic for constructing these names belongs in `internal/storage`.

### 4. Implement the application and HTTP request handling (`internal/app`)

Define `Application`, which owns the methods for receiving events and serving the frontend, and `Dependencies`, which holds the components it needs:

- A `Store` for saving events.
- An `fs.FS` for accessing frontend files. This is Go's standard filesystem interface; it lets the application access files without knowing whether they are embedded in the binary or supplied by a test.

Passing these components into `Application` is known as dependency injection. It means tests can supply substitutes without modifying the application logic.

#### Receiving telemetry events

For `POST /`, the application must check the content type, decode the JSON body, validate it using `internal/event`, pass the validated envelope (including `params`) to `Store`, and return the appropriate HTTP response. Do not pass the raw POST body to storage.

After a successful save, print that validated event (compact JSON) to stdout. This is temporary local scaffolding so a manual POST against a locally running function can be inspected in the terminal — the in-memory store has no other way to browse its contents. Keep the print as simple as possible; do not add a logging framework or configuration. Remove it in Phase 3 before the first Google Cloud deploy (see [developer-workflow.md](../developer-workflow.md)).

The request body must contain one envelope object as described in [event-schema.md](../event-schema.md).

The HTTP contract is:

| Situation | HTTP response |
| --- | --- |
| Valid event successfully saved | `204 No Content` (empty body) |
| Incorrect content type | `415 Unsupported Media Type` |
| Invalid JSON or invalid envelope | `400 Bad Request` |
| Storage failure | `500 Internal Server Error` |

Accept `Content-Type: application/json`, including the optional `; charset=utf-8` suffix. No response body is required for these statuses in Phase 2. The ingest endpoint is `POST /` only.

#### Serving the frontend

Use `fs.Sub` and `http.FileServer` to serve frontend files from the filesystem supplied through `Dependencies`.

Support `GET` and `HEAD` on `/` and real asset paths. For frontend routes, support SPA (single-page application) fallback to `index.html`. This allows the frontend to handle its own routes even when there is no corresponding physical file.

However, missing assets that should return `404 Not Found` must not fall back to `index.html`. For example, a request for a nonexistent JavaScript file should not return the HTML page instead.

Unsupported HTTP methods must return `405 Method Not Allowed` with an `Allow` header, following [high-level-architecture.md](../high-level-architecture.md).

Tests for frontend serving should supply a small in-memory filesystem, such as `fstest.MapFS`, rather than reading `static/` or requiring Vite.

### 5. Assemble the application in `function.go`

Create the root-package entry point for the Go Functions Framework. This file is the **composition root**: the place where the application's components are assembled and connected.

It must:

1. Embed the generated frontend files using `//go:embed static`.
2. Construct the runtime (production) `Dependencies`, using the in-memory store and embedded filesystem.
3. Construct `Application` with those dependencies.
4. Register one HTTP handler with the Functions Framework.
5. Delegate incoming requests immediately to `Application`.

The Functions Framework provides the HTTP entry point that will eventually be used in Google Cloud. It can also be run locally.

Keep request handling out of `function.go`. It must not perform validation or call storage itself. Those responsibilities belong to the internal packages, so they can be tested independently of the framework.

Compile the root package only through the full-verify or future deploy path, after Make has staged `static/`.

### 6. Build the minimal Svelte frontend (`ui/`)

Create a minimal Svelte + Vite application displaying Hello World. It needs a package manifest, lockfile, `.nvmrc` pinning Node 22, Vite configuration, `index.html`, source entry point and `App.svelte`.

The frontend build works as follows:

```text
Svelte source in ui/
        ↓ Vite production build
ui/dist/
        ↓ Make copies the built files
static/
        ↓ Go embeds the files
Compiled Go application
        ↓ HTTP GET
Webpage in the browser
```

Go does not need to understand Svelte: it simply serves the HTML, JavaScript and other files that Vite produces.

There is no telemetry dashboard or analysis functionality in this phase. Hello World is enough to prove that the frontend build and serving mechanism work.

### 7. Write automated tests and verify the whole application

The normal backend test command is `go test ./internal/...`. It must run without Google Cloud, a frontend build or a staged `static/` directory.

Prefer testing `Application` directly with a memory `Store` and injected filesystem rather than starting a live Functions Framework process.

The tests should demonstrate the following behaviour, covering the contracts defined in the preceding steps:

- Valid POST requests are accepted and the validated events appear in the memory store.
- Bad envelopes and malformed requests are rejected with the specified HTTP statuses.
- Stored events have the required JSON format and object names.
- GET serves UI bytes from an injected `fs.FS`, with appropriate handling of real assets, SPA fallback and missing assets.
- HEAD and unsupported methods follow the HTTP contract.

Full verification must additionally build the actual Svelte Hello World frontend, stage it in `static/` and run Go checks for the whole module, including the root package and its embed.

A small local Functions Framework run is optional; it is not a prerequisite for the everyday automated tests.

## Out of scope

Do not introduce any of the following in Phase 2:

- Google Cloud, Cloud Function deployment or GCS.
- Changes to `dxact-draw` or `dxact-wasm`.
- An analysis model, product-question recipes or an observability UI beyond Hello World.
- Rate limits, request-size limits or production hardening.
- Validation of the individual parameter fields for each Now event (validate the common envelope only).

## Done when

- [ ] The Go application accepts valid POST events and stores them in the in-memory store in the required format.
- [ ] Invalid events are rejected and the HTTP response contract is satisfied.
- [ ] GET serves the Svelte Hello World webpage and its assets.
- [ ] Backend tests demonstrate event ingest, rejection and UI serving without cloud infrastructure or a frontend build.
- [ ] Full verification builds and stages the frontend and successfully builds/tests the whole Go module, including the embedded frontend.

This is the same completion criterion as Phase 2 in [developer-workflow.md](../developer-workflow.md): a working local ingest path and trivial UI, with automated tests and no cloud dependencies.
