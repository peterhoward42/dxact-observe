# Technology choices

These are fixed constraints simply to reflect preferences and simplify the design
space.

- Use the Go programming lanugage for the cloud function
- Use Google's Cloud Function Framework package
- Assume a unix-like dev and runtime environment

# Repository layout

- One repository with two top-level directories:
	- `ingestion/` — the Go backend
	- `ui/` — a separate Svelte application, organised comparably with
	  dxact-draw

# To serve the UI

- Use Svelte to compose the UI as a SPA under `ui/`.
- The intended arrangement is that the Cloud Function serves the built static
  assets (HTML, JavaScript, CSS) as the response to GET, potentially embedding
  them with Go's `embed.FS`.
- This is feasible pending validation — see high-level-architecture.md
  (single endpoint; Phase 2 proof of concept locally then on Google Cloud).
