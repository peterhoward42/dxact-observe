# High Level Architecture

- An https server manifested as a single URL endpoint
- A POST handler to receive events
- A GET handler to serve the page(s) of the observation UI
- A family of GET paths mapped to a set of screens or tabs that show one 
  of the dedicated analysis recipes that we hard code

## Single endpoint — feasible, pending validation

- Google's documentation confirms that a Go Cloud Function can handle both
  GET and POST requests.
- Go can also serve a Svelte application's compiled HTML, JavaScript and CSS
  assets, potentially embedding them in the binary using `embed.FS`.
- We are therefore fairly confident this arrangement will work. Deployment
  integration, static asset routing and performance still need verifying.
- Phase 2 should include a small proof of concept that validates the complete
  arrangement, initially locally and then on Google Cloud, before committing
  to the architecture.

## HTTP handling and application logic

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

## Consuming telemetry events

- Incoming events should be stored in GCS
- Use a single bucket named `events`
- Use a new blob to model and accomodate events that carry events with a
  particular time window (UTC)
- Incoming events should be serialised to single line JSON and appended as one
  line to one of these blobs
- Each blob should have a filename that encodes its time envelope slot using a
  well known convention for that.
- Object names are time-based. Do not encode event type or user identifier into
  object names.

### Open risk — concurrent append

- GCS does not support straightforward concurrent appends. If two Cloud
  Function instances update the same object at once, one write can overwrite
  the other and events are lost.
- This risk is recorded; the storage shape above is unchanged for now. Whether
  to switch to one object per event under a time-keyed prefix (combining them
  when building the report model) remains open.

## Serving the UI

- We will define a set of query shapes over the event store
- This will start as a small set and will change significantly over time
- The results of each such query shape / recipe will be displayed by whatever
  layout serves that query well
