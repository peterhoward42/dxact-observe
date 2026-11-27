package event_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/peterhoward42/dxact-observe/internal/event"
)

// This file pins Parse against the common envelope rules in event-schema.md:
// required fields, Z-suffixed UTC time, plain browser UUID form, schema
// version 1, and params as a JSON object. Catalogue-specific params are not
// checked here.

// TestParseAcceptsValidEnvelope checks that a full valid payload becomes an
// Envelope with the expected field values.
func TestParseAcceptsValidEnvelope(t *testing.T) {
	t.Parallel()

	payload := validPayload()
	body := mustMarshal(t, payload)

	got, err := event.Parse(body)
	if err != nil {
		t.Fatalf("Parse valid envelope: %v", err)
	}

	wantTime := mustParseTime(t, payload["time"].(string))
	if !got.Time.Equal(wantTime) {
		t.Fatalf("Time: got %v, want %v", got.Time, wantTime)
	}
	if got.Type != payload["type"] {
		t.Fatalf("Type: got %q, want %q", got.Type, payload["type"])
	}
	if got.BrowserUUID != payload["browser_uuid"] {
		t.Fatalf(
			"BrowserUUID: got %q, want %q",
			got.BrowserUUID,
			payload["browser_uuid"],
		)
	}
	if got.SchemaVersion != event.SchemaVersion {
		t.Fatalf(
			"SchemaVersion: got %d, want %d",
			got.SchemaVersion,
			event.SchemaVersion,
		)
	}
	wantParams := mustMarshal(t, payload["params"])
	if string(got.Params) != string(wantParams) {
		t.Fatalf(
			"Params: got %s, want %s",
			got.Params,
			wantParams,
		)
	}
}

// TestParseAcceptsEmptyParamsObject checks that params may be {}.
func TestParseAcceptsEmptyParamsObject(t *testing.T) {
	t.Parallel()

	payload := validPayload()
	payload["params"] = map[string]any{}
	body := mustMarshal(t, payload)

	got, err := event.Parse(body)
	if err != nil {
		t.Fatalf("Parse empty params: %v", err)
	}
	if string(got.Params) != "{}" {
		t.Fatalf("Params: got %s, want {}", got.Params)
	}
}

// TestParseAcceptsTimeWithoutFractionalSeconds checks that whole-second Z
// timestamps are accepted as well as fractional ones.
func TestParseAcceptsTimeWithoutFractionalSeconds(t *testing.T) {
	t.Parallel()

	payload := validPayload()
	payload["time"] = "2026-11-24T08:45:12Z"
	body := mustMarshal(t, payload)

	got, err := event.Parse(body)
	if err != nil {
		t.Fatalf("Parse whole-second time: %v", err)
	}
	want := mustParseTime(t, "2026-11-24T08:45:12Z")
	if !got.Time.Equal(want) {
		t.Fatalf("Time: got %v, want %v", got.Time, want)
	}
}

// TestParseRejectsInvalidEnvelopes table-tests the rejection paths: missing
// or invalid fields, unknown keys, non-object params, and malformed or
// multi-value JSON. Each case names the substring expected in the error.
func TestParseRejectsInvalidEnvelopes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		errContains string
		mutate      func(map[string]any)
		body        func(t *testing.T) []byte
	}{
		{
			name:        "missing time",
			errContains: "time is required",
			mutate: func(p map[string]any) {
				delete(p, "time")
			},
		},
		{
			name:        "missing type",
			errContains: "type is required",
			mutate: func(p map[string]any) {
				delete(p, "type")
			},
		},
		{
			name:        "missing browser_uuid",
			errContains: "browser_uuid is required",
			mutate: func(p map[string]any) {
				delete(p, "browser_uuid")
			},
		},
		{
			name:        "missing schema_version",
			errContains: "schema_version is required",
			mutate: func(p map[string]any) {
				delete(p, "schema_version")
			},
		},
		{
			name:        "missing params",
			errContains: "params is required",
			mutate: func(p map[string]any) {
				delete(p, "params")
			},
		},
		{
			name:        "unknown top-level key",
			errContains: "unknown field",
			mutate: func(p map[string]any) {
				p["extra"] = "nope"
			},
		},
		{
			name:        "empty type",
			errContains: "type must be non-empty",
			mutate: func(p map[string]any) {
				p["type"] = ""
			},
		},
		{
			name:        "time with +00:00 offset",
			errContains: "time must be RFC 3339 UTC with a Z suffix",
			mutate: func(p map[string]any) {
				p["time"] = "2026-11-24T08:45:12.345+00:00"
			},
		},
		{
			name:        "time with non-UTC offset",
			errContains: "time must be RFC 3339 UTC with a Z suffix",
			mutate: func(p map[string]any) {
				p["time"] = "2026-11-24T09:45:12.345+01:00"
			},
		},
		{
			name:        "malformed time",
			errContains: "time must be RFC 3339 UTC with a Z suffix",
			mutate: func(p map[string]any) {
				p["time"] = "not-a-timestamp"
			},
		},
		{
			name:        "browser_uuid with braces",
			errContains: "browser_uuid must be a UUID in 8-4-4-4-12 form",
			mutate: func(p map[string]any) {
				p["browser_uuid"] =
					"{8e6a4c1d-5f29-4e83-9e10-42bd735af0c2}"
			},
		},
		{
			name:        "browser_uuid with urn prefix",
			errContains: "browser_uuid must be a UUID in 8-4-4-4-12 form",
			mutate: func(p map[string]any) {
				p["browser_uuid"] =
					"urn:uuid:8e6a4c1d-5f29-4e83-9e10-42bd735af0c2"
			},
		},
		{
			name:        "browser_uuid not a UUID",
			errContains: "browser_uuid must be a UUID in 8-4-4-4-12 form",
			mutate: func(p map[string]any) {
				p["browser_uuid"] = "not-a-uuid"
			},
		},
		{
			name:        "schema_version zero",
			errContains: "schema_version must be 1",
			mutate: func(p map[string]any) {
				p["schema_version"] = 0
			},
		},
		{
			name:        "schema_version two",
			errContains: "schema_version must be 1",
			mutate: func(p map[string]any) {
				p["schema_version"] = 2
			},
		},
		{
			name:        "schema_version string",
			errContains: "invalid event JSON",
			mutate: func(p map[string]any) {
				p["schema_version"] = "1"
			},
		},
		{
			name:        "params null",
			errContains: "params must be a JSON object",
			mutate: func(p map[string]any) {
				p["params"] = nil
			},
		},
		{
			name:        "params array",
			errContains: "params must be a JSON object",
			mutate: func(p map[string]any) {
				p["params"] = []any{"x"}
			},
		},
		{
			name:        "params string",
			errContains: "params must be a JSON object",
			mutate: func(p map[string]any) {
				p["params"] = "nope"
			},
		},
		{
			name:        "params number",
			errContains: "params must be a JSON object",
			mutate: func(p map[string]any) {
				p["params"] = 1
			},
		},
		{
			name:        "top-level array",
			errContains: "invalid event JSON",
			body: func(t *testing.T) []byte {
				t.Helper()
				return mustMarshal(t, []any{validPayload()})
			},
		},
		{
			// A trailing empty object exercises the dedicated
			// "trailing data" path (a second full envelope would
			// fail earlier as an unknown-field decode error).
			name:        "trailing second value",
			errContains: "trailing data",
			body: func(t *testing.T) []byte {
				t.Helper()
				first := mustMarshal(t, validPayload())
				return append(append(first, '\n'), '{', '}')
			},
		},
		{
			name:        "malformed JSON",
			errContains: "invalid event JSON",
			body: func(t *testing.T) []byte {
				t.Helper()
				return []byte(`{"time":`)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var body []byte
			switch {
			case tc.body != nil:
				body = tc.body(t)
			default:
				payload := validPayload()
				tc.mutate(payload)
				body = mustMarshal(t, payload)
			}

			_, err := event.Parse(body)
			if err == nil {
				t.Fatalf("Parse accepted invalid envelope")
			}
			if !strings.Contains(err.Error(), tc.errContains) {
				t.Fatalf(
					"error %q does not contain %q",
					err.Error(),
					tc.errContains,
				)
			}
		})
	}
}

// validPayload returns a new valid envelope map for one test to mutate.
func validPayload() map[string]any {
	return map[string]any{
		"time":          "2026-11-24T08:45:12.345Z",
		"type":          "Application opened",
		"browser_uuid":  "8e6a4c1d-5f29-4e83-9e10-42bd735af0c2",
		"schema_version": event.SchemaVersion,
		"params": map[string]any{
			"face":    "desktop",
			"arrival": "plain",
		},
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
