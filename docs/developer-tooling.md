# Developer tooling

- Developer tooling should be based on Makefiles — following styles and
  examples used by the dxact-wasm repo where useful.
- Make targets may delegate out to Python scripts — again following precedents
  in dxact-wasm.

## Make owns integrated build and deploy

**Go and Make are authoritative.** The UI can be developed using normal
Svelte/Vite tooling, but the top-level Makefile owns the integrated build and
deployment:

1. Install or verify pinned frontend dependencies (for example, `npm ci` when
   appropriate).
2. Run the production Svelte/Vite build in `ui/`.
3. Copy the resulting `ui/dist/` files into `static/` within the Go source
   tree.
4. Run Go checks/tests as required.
5. Deploy the Cloud Function **from source**, including the generated
   `static/` directory.

A `//go:embed static` directive in the Go source incorporates the files when
Google Cloud's build environment compiles the Go binary. `go:embed` is a
compile-time facility, not a replacement for Make orchestration. In
particular, embedding cannot directly reach into an arbitrary directory
outside the Go package tree, so staging the assets under the Go source is
deliberate.

The generated `static/` contents would normally be excluded from Git, **but
must not be excluded from the deployment upload**. Check `.gcloudignore`
explicitly, especially if it inherits patterns from `.gitignore`. The
build/deploy targets should guarantee that fresh assets exist before source
deployment starts.

## Reproducibility

- Commit the npm lockfile.
- Use a consistent Node version.
- Ensure Make stages the assets before deployment.
