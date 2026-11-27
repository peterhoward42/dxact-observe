package storage_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/peterhoward42/dxact-observe/internal/event"
)

// Shared helpers for storage package tests.

// validEnvelope returns a fresh validated envelope for one test.
func validEnvelope(t *testing.T) event.Envelope {
	t.Helper()
	return event.Envelope{
		Time:          mustParseTime(t, "2026-11-24T08:45:12.345Z"),
		Type:          "Application opened",
		BrowserUUID:   "8e6a4c1d-5f29-4e83-9e10-42bd735af0c2",
		SchemaVersion: event.SchemaVersion,
		Params: mustMarshal(t, map[string]any{
			"face":    "desktop",
			"arrival": "plain",
		}),
	}
}

func mustMarshal(t *testing.T, value any) []byte {
	t.Helper()
	body, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return body
}

func mustParseTime(t *testing.T, value string) time.Time {
	t.Helper()
	when, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		t.Fatalf("time.Parse: %v", err)
	}
	return when.UTC()
}
