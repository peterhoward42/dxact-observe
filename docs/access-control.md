#  Access control

## Event ingestion

- A public endpoint
- Guarded only by the silent rejection of suspicious looking payloads
- Payload validation alone will not protect against abuse such as large
  volumes of otherwise valid requests. Rate limiting, request size limits
  and cloud-level quotas may need consideration before deployment; they
  are not decided or implemented yet.

## UI serving

- A public endpoint

## Storage access

- The GCS bucket must require authentication to access 
- The requisite secrets will be defined in the Google Cloud Project's defined
  environment variables
