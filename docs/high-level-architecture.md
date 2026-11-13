# High Level Architecture

- An https server manifested as a single URL endpoint
- A POST handler to receive events
- A GET handler to serve the page(s) of the observation UI
- A family of GET paths mapped to a set of screens or tabs that show one 
  of the dedicated analysis recipes that we hard code

## Consuming telemetry events

- Incoming events should be stored in GCS
- Use a single bucket
- Use a new blob to model and accomodate events that carry events with a
  particular time window (UTC)
- Incoming events should be serialised to single line JSON and appended as one
  line to one of these blobs
- Each blob should have a filename that encodes its time envelope slot using a
  well known convention for that.


## Serving the UI

- We will define a set of query shapes over the event store
- This will start as a small set and will change significantly over time
- The results of each such query shape / recipe will be displayed by whatever
  layout serves that query well
