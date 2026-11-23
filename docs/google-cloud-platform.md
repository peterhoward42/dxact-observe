# Google Cloud Platform

- Use a newly minted GCP project
- Define it to use a Google Cloud Function using the Go language, with
  **source** deployment — including the generated `static/` directory that
  holds the embedded Svelte build output.
- Use the Google Cloud Function Framework package
- No separately managed container image or frontend hosting service for the
  initial design; revisit separate static hosting or a CDN only if traffic,
  caching, or operational requirements justify it (see
  high-level-architecture.md).
- Assume the necessary Google Cloud project and credentials will eventually
  exist. Project-specific values can start as constants or configuration
  variables; actual deployment configuration can come later and need not block
  application design.
- The generated `static/` tree must be present in the deployment upload even
  if it is excluded from Git. Check `.gcloudignore` explicitly, especially if
  it inherits patterns from `.gitignore` — see developer-tooling.md.
- Cold starts: a function scaled to zero may delay the initial page load;
  minimum instances are an option if this matters. Each static asset request
  invokes the function; that cost is accepted at current expected volumes.
- It is often inconvenient to code it to use the option to also run it
  locally.
  	- We don't really need that
	- With properly structured code and adherence to proper separation of
	  concerns, local tests at unit and a degree of integration testing should
	  be sufficient.
