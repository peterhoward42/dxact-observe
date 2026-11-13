# Technology choices

These are fixed constraints simply to reflect preferences and simplify the design
space.

- Use the Go programming lanugage for the cloud function
- Use Google's Cloud Function Framework package
- Assume a unix-like dev and runtime environment

# To serve the UI

- I am not sure about this
- I would prefer to use Svelte to compose the UI, but I'm not sure how to
  organise the code for this and have it integrated into the cloud function
  server.
	- Maybe you can define a directory in the repo that IS the Svelte app
	- Design it as a SPA
	- And have the cloud function serve the SPA as the response to the GET
	  handler?
