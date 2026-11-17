# Google Cloud Platform

- Use a newly minted GCP project
- Define it to use a Google Cloud Function using the Go language, with "source"
  deployment.
- Use the Google Cloud Function Framework package
- Assume the necessary Google Cloud project and credentials will eventually
  exist. Project-specific values can start as constants or configuration
  variables; actual deployment configuration can come later and need not block
  application design.
- It is often inconvenient to code it to use the option to also run it 
  locally.
  	- We don't really need that
	- With properly structed code and adherence to proper separation of 
	  concerns, local tests at unit and a degree of integration testing should be
	  sufficient.
