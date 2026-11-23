# Technology choices

These are fixed constraints simply to reflect preferences and simplify the design
space.

- Use the Go programming lanugage for the cloud function
- Use Google's Cloud Function Framework package
- Assume a unix-like dev and runtime environment

# Working decision — function and UI

Proceed with **one monorepo, one Go HTTP Cloud Function, an embedded production
Svelte SPA, and a Make-driven source deployment**. Keep the initial
implementation simple and revisit separate frontend hosting only if traffic,
caching, or operational requirements justify it.

- No separately managed container image or frontend hosting service for now.
- Keep the Go Functions Framework and source-deployment workflow.
- Svelte owns frontend source and compilation; Go/Make own assembly,
  verification, and deployment.
- The browser runs the Svelte application; Go serves its compiled assets and
  implements the server-side endpoints. UI requests to the backend can use the
  same origin, avoiding cross-origin configuration for that interaction.
- HTTP routing and trade-offs: see high-level-architecture.md.
- Build/deploy pipeline: see developer-tooling.md.

# Repository layout

This project is intentionally a **monorepo**, because it is small and closely
integrated. Larger projects may use separate sibling repositories, but that
convention is not needed here.

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
There is no separate top-level `ingestion/` directory; Go lives at the repo
root alongside `ui/` and the generated `static/` tree.

# To serve the UI

- Use Svelte to compose the UI as a SPA under `ui/`, organised comparably with
  dxact-draw where that helps.
- The Cloud Function serves the built static assets (HTML, JavaScript, CSS) as
  the response to GET, embedding them with Go's `//go:embed static` after Make
  stages `ui/dist/` into `static/`.
- Phase 2 proves the decided shape locally (POST plus serving the UI from one
  process). The first Google Cloud deploy waits until Phase 3 with GCS — see
  developer-workflow.md and high-level-architecture.md.
