# Processing overview assumptions

- Ingestion is concerned only with:
	- decoding and validating events
	- serialising each event as compact single-line JSON
	- writing one new GCS object per event under
	  `events/YYYY-MM-DD/<event-uuid>.json` (minting a fresh event UUID for
	  the object name; see high-level-architecture.md)

- Queries should not run over raw events
- Instead the repo should define an explicit data model
- Each query should fetch the raw events from storage, so as to build an in
  memory instance of this data model
- The Cloud Function (or equivalent loader) lists and reads stored events,
  parses them into that model, groups them by browser UUID, and orders them by
  event timestamp for analysis and UI display
- Ordering uses the UTC event time in the payload, not object upload or
  creation time. Network delays and retries can mean events arrive out of
  storage order
- The queries should be defined to run over the data model
- Thus the query logic can have tests written for them without any coupling to
  raw events or the blob storage
- The data model need not be the same shape as the raw data
- The data model may well benefit from being composed of maps and indexes for
  example that make the queries easier to evaluate

## Loading into the UI

- The observability UI should build the in-memory data model when it starts and
  keep that instance unchanged for the session.
- Reloading the page retrieves the latest telemetry and rebuilds the model.
- How the UI retrieves stored events and constructs the model remains open.
