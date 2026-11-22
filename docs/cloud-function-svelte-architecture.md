# Go Cloud Function with an Embedded Svelte UI

## Purpose

Build a small, self-contained Google Cloud Function that both receives events and serves a Svelte single-page application (SPA). Keep the existing Go Functions Framework and source-deployment workflow, without introducing a separately managed container image or frontend hosting service.

## Proposed architecture

A single HTTP Cloud Function exposes one origin:

- **POST `/`** receives events, validates and processes them, and writes the appropriate data to Google Cloud storage services.
- **GET `/`** serves the Svelte application's `index.html`.
- **GET `/assets/...`** (and other static asset paths) serves the compiled JavaScript, CSS, and other assets.
- **GET on SPA navigation paths**, if client-side routing is used, falls back to `index.html` when the path is not a real asset.

The browser runs the Svelte application; Go only serves its compiled assets and implements the server-side endpoints. Requests from the UI to the backend can use the same origin, avoiding cross-origin configuration for that interaction.

## Repository layout

This project is intentionally a **monorepo**, because it is small and closely integrated. Larger projects may use separate sibling repositories, but that convention is not needed here.

```text
project/
├── Makefile                # Authoritative build/deploy orchestration
├── go.mod
├── go.sum
├── function.go            # Functions Framework HTTP entry point
├── static/                # Generated Svelte build output, embedded by Go
│   └── ...
└── ui/
    ├── package.json
    ├── package-lock.json
    ├── vite.config.js
    ├── index.html
    └── src/
```

The precise names and package layout can change as the implementation develops.

## Build and deployment model

**Go and Make are authoritative.** The UI can be developed using normal Svelte/Vite tooling, but the top-level Makefile owns the integrated build and deployment:

1. Install or verify pinned frontend dependencies (for example, `npm ci` when appropriate).
2. Run the production Svelte/Vite build in `ui/`.
3. Copy the resulting `ui/dist/` files into `static/` within the Go source tree.
4. Run Go checks/tests as required.
5. Deploy the Cloud Function **from source**, including the generated `static/` directory.

A `//go:embed static` directive in the Go source incorporates the files when Google Cloud's build environment compiles the Go binary. `go:embed` is a compile-time facility, not a replacement for Make orchestration. In particular, embedding cannot directly reach into an arbitrary directory outside the Go package tree, so staging the assets under the Go source is deliberate.

The generated `static/` contents would normally be excluded from Git, **but must not be excluded from the deployment upload**. Check `.gcloudignore` explicitly, especially if it inherits patterns from `.gitignore`. The build/deploy targets should guarantee that fresh assets exist before source deployment starts.

## HTTP handling

Register one HTTP handler with the Go Functions Framework. Dispatch by HTTP method and path:

- `POST /`: event ingestion.
- `GET /` and `HEAD /`: serve the SPA entry page.
- `GET /assets/...` and `HEAD /assets/...`: serve embedded static files.
- Other GET navigation paths: optionally return `index.html` for client-side routing, without hiding missing asset errors.
- Unsupported methods: return `405 Method Not Allowed` with an appropriate `Allow` header.

Use `embed.FS`, `fs.Sub`, and `http.FileServer` as the starting point for serving the compiled frontend, with explicit SPA fallback behavior if needed.

## Rationale

- **Minimal infrastructure:** one deployed function handles the UI and event ingestion.
- **Familiar operational model:** preserves Go, the Functions Framework, Linux-like development environments, Make automation, and Google Cloud source deployment.
- **One coherent release:** frontend and backend are built and deployed together from a single repository revision.
- **Simple browser integration:** UI requests can reach the backend at the same origin.
- **Small-project fit:** avoids the coordination overhead of multiple repositories and independent hosting/deployment pipelines.
- **Clear ownership:** Svelte owns frontend source and compilation; Go/Make owns assembly, verification, and deployment.

## Trade-offs and implementation checks

- **Cold starts:** a function scaled to zero may delay the initial page load; minimum instances are an option if this matters.
- **Static serving cost:** each asset request invokes the function. For substantially higher traffic, separate static hosting or a CDN may become preferable.
- **Caching:** hashed Vite assets can receive long-lived cache headers; `index.html` should be refreshed more conservatively.
- **Security:** if public GET access requires an unauthenticated function, POST ingestion needs its own authentication or request verification as appropriate. HTTP method separation alone is not a security boundary.
- **Deployment coupling:** a UI-only change still requires redeploying the function; for this small project, the simplicity of one release is likely worth it.
- **Routing:** static asset requests should return actual 404s when missing, while intentional SPA navigation routes can fall back to `index.html`.
- **Reproducibility:** commit the npm lockfile, use a consistent Node version, and ensure Make stages the assets before deployment.

## Working decision

Proceed with **one monorepo, one Go HTTP Cloud Function, an embedded production Svelte SPA, and a Make-driven source deployment**. Keep the initial implementation simple and revisit separate frontend hosting only if traffic, caching, or operational requirements justify it.
