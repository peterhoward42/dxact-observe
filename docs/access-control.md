#  Access control

## Event ingestion

- A public endpoint
- Guarded only by the silent rejection of suspicious looking payloads

## UI serving

- A public endpoint

## Storage access

- The GCS bucket must require authentication to access 
- The requisite secrets will be defined in the Google Cloud Project's defined
  environment variables
