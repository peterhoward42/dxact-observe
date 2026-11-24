// Package event defines and validates the common telemetry event envelope
// used by ingestion and subsequent analysis.
// It validates the fields shared by all event types, but does not
// validate the contents of the event-specific params object.
package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SchemaVersion is the only envelope schema_version accepted today.
// Versioning allows the schema to evolve while distinguishing events recorded
// under different versions.
const SchemaVersion = 1

// Envelope is one validated telemetry event after ingest checks the common
// fields. Params is the event-specific JSON object; its keys are not
// inspected at this layer.
type Envelope struct {
	Time          time.Time
	Type          string
	BrowserUUID   string
	SchemaVersion int
	Params        json.RawMessage
}

// Parse decodes a single JSON object representing a telemetry event,
// validates it against the envelope schema, and returns an Envelope.
// It returns an error if the JSON is malformed or does not conform
// to the schema.
// Event-specific parameter contents are not validated.
func Parse(data []byte) (Envelope, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	// Use pointer fields to distinguish missing or null values
	// from explicitly supplied zero values.
	var raw struct {
		Time          *string         `json:"time"`
		Type          *string         `json:"type"`
		BrowserUUID   *string         `json:"browser_uuid"`
		SchemaVersion *int            `json:"schema_version"`
		Params        json.RawMessage `json:"params"`
	}
	if err := dec.Decode(&raw); err != nil {
		return Envelope{}, fmt.Errorf("invalid event JSON: %w", err)
	}
	// The body must be exactly one JSON value; only trailing whitespace is allowed.
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return Envelope{}, fmt.Errorf(
				"invalid event JSON: trailing data",
			)
		}
		return Envelope{}, fmt.Errorf("invalid event JSON: %w", err)
	}

	// Validate the presence and value of each required envelope field.
	if raw.Time == nil {
		return Envelope{}, fmt.Errorf("time is required")
	}
	when, err := parseEventTime(*raw.Time)
	if err != nil {
		return Envelope{}, err
	}

	if raw.Type == nil {
		return Envelope{}, fmt.Errorf("type is required")
	}
	if *raw.Type == "" {
		return Envelope{}, fmt.Errorf("type must be non-empty")
	}

	if raw.BrowserUUID == nil {
		return Envelope{}, fmt.Errorf("browser_uuid is required")
	}
	if err := requireBrowserUUID(*raw.BrowserUUID); err != nil {
		return Envelope{}, err
	}

	if raw.SchemaVersion == nil {
		return Envelope{}, fmt.Errorf("schema_version is required")
	}
	if *raw.SchemaVersion != SchemaVersion {
		return Envelope{}, fmt.Errorf(
			"schema_version must be %d",
			SchemaVersion,
		)
	}

	params, err := requireParamsObject(raw.Params)
	if err != nil {
		return Envelope{}, err
	}

	return Envelope{
		Time:          when,
		Type:          *raw.Type,
		BrowserUUID:   *raw.BrowserUUID,
		SchemaVersion: *raw.SchemaVersion,
		Params:        params,
	}, nil
}

// requireBrowserUUID accepts only the standard 8-4-4-4-12 hex form.
// google/uuid.Validate also allows braces and urn:uuid: prefixes; those are
// rejected here so the wire form matches the event schema.
func requireBrowserUUID(value string) error {
	if len(value) != 36 {
		return fmt.Errorf(
			"browser_uuid must be a UUID in 8-4-4-4-12 form",
		)
	}
	if err := uuid.Validate(value); err != nil {
		return fmt.Errorf(
			"browser_uuid must be a UUID in 8-4-4-4-12 form: %w",
			err,
		)
	}
	return nil
}

// parseEventTime accepts an RFC 3339 UTC timestamp with a Z suffix.
// Fractional seconds are allowed.
func parseEventTime(value string) (time.Time, error) {	
	// RFC3339Nano also accepts numeric timezone offsets, but the event
	// schema requires timestamps to use the Z suffix.
	if !strings.HasSuffix(value, "Z") {
		return time.Time{}, fmt.Errorf(
			"time must be RFC 3339 UTC with a Z suffix",
		)
	}
	when, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf(
			"time must be RFC 3339 UTC with a Z suffix: %w",
			err,
		)
	}
	return when.UTC(), nil
}

// requireParamsObject validates and returns params as a JSON object,
// including an empty object ({}), with surrounding whitespace removed.
// Event-specific contents are not validated.
func requireParamsObject(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("params is required")
	}
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return nil, fmt.Errorf("params must be a JSON object")
	}
	// Decode into a temporary map to verify the JSON object's validity
	// while retaining the original JSON for subsequent processing.
	var object map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &object); err != nil {
		return nil, fmt.Errorf("params must be a JSON object: %w", err)
	}
	return json.RawMessage(trimmed), nil
}
