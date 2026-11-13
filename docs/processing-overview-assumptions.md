# Processing overview assumptions

- Ingestion is concerned only with:
	- decoding and validating events 
	- choosing and creating a new time envelope blob when necessary
	- serialising them in a human readable and compact way
		- probably single line JSON

- Queries should not run over raw events
- Instead the repo should define an explicit data model
- Each query should fetch the raw events from storage, so as to build an in
  memory instance of this data model
- The queries should be defined to run over the data model
- Thus the query logic can have tests written for them without any coupling to
  raw events or the blob storage
- The data model need not be the same shape as the raw data
- The data model may well benefit from being composed of maps and indexes for
  example that make the queries easier to evaluate
