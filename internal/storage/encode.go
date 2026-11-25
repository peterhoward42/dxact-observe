package storage

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/peterhoward42/dxact-observe/internal/event"
)

// storedEnvelope defines the JSON representation of an event for storage.
// The time field comes first to make single-line JSON records easier to scan.
type storedEnvelope struct {
	Time          string          `json:"time"`
	Type          string          `json:"type"`
	BrowserUUID   string          `json:"browser_uuid"`
	SchemaVersion int             `json:"schema_version"`
	Params        json.RawMessage `json:"params"`
}

// EncodeEnvelope re-encodes a validated envelope as compact single-line JSON
// with time first. 
// The caller is responsible for validating the envelope before encoding.
func EncodeEnvelope(envelope event.Envelope) ([]byte, error) {
	body, err := json.Marshal(storedEnvelope{
		Time:          formatEventTime(envelope.Time),
		Type:          envelope.Type,
		BrowserUUID:   envelope.BrowserUUID,
		SchemaVersion: envelope.SchemaVersion,
		Params:        envelope.Params,
	})
	if err != nil {
		return nil, fmt.Errorf("encode envelope: %w", err)
	}
	return body, nil
}

// formatEventTime formats a timestamp as RFC3339 in UTC, with a Z suffix.
// Fractional seconds are included only when needed, without trailing zeros.
func formatEventTime(when time.Time) string {
	return when.UTC().Format(time.RFC3339Nano)
}
