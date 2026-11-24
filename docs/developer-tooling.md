# Developer tooling

- Developer tooling should be based on Makefiles — following styles and
  examples used by the dxact-wasm repo where useful.
- Make targets may delegate out to Python scripts — again following precedents
  in dxact-wasm.

## Make owns integrated build and deploy

**Go and Make are authoritative.** The UI can be developed using normal
Svelte/Vite tooling, but the top-level Makefile owns the integrated build and
deployment.

Everyday backend work does not need a frontend build. A fast Make target runs
Go tests for the internal packages only (`./internal/...`). Those tests never
require `static/` or Vite. UI-serving unit tests inject a small in-memory
`fs.FS` through `Dependencies`.

The full application path is separate:

1. Install or verify pinned frontend dependencies (for example, `npm ci` when
   appropriate).
2. Run the production Svelte/Vite build in `ui/`.
3. Copy the resulting `ui/dist/` files into `static/` within the Go source
   tree.
4. Run Go checks/tests for the whole module (including the root package that
   embeds `static/`).
5. Deploy the Cloud Function **from source**, including the generated
   `static/` directory.

`//go:embed static` lives only in the root package (alongside `function.go`).
Production wiring passes that filesystem into `Dependencies`. `go:embed` is a
compile-time facility, not a replacement for Make orchestration. Embedding
cannot reach an arbitrary directory outside the Go package tree, so staging
under the Go source is deliberate.

The generated `static/` tree is gitignored. Nothing under `static/` is
committed as a placeholder. A fresh clone can run everyday backend tests with
no staged assets. Full verify and deploy targets must stage fresh assets before
any Go build of the root package or source upload. Check `.gcloudignore`
explicitly so `static/` is not excluded from the deployment upload if that
ignore file inherits patterns from `.gitignore`.

## Reproducibility

- Commit the npm lockfile.
- Pin Node.js with `ui/.nvmrc` (Node **22** for Phase 2). Make’s UI build
  path should use that version (for example via `nvm use` or an equivalent
  check).
- Ensure Make stages the assets before full verification and deployment.
