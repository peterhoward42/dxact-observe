package storage_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/peterhoward42/dxact-observe/internal/storage"
)

func TestEncodeEnvelopeIsCompactJSONWithTimeFirst(t *testing.T) {
	t.Parallel()

	envelope := validEnvelope(t)
	got, err := storage.EncodeEnvelope(envelope)
	if err != nil {
		t.Fatalf("EncodeEnvelope: %v", err)
	}
	if bytes.Contains(got, []byte{'\n'}) {
		t.Fatalf("EncodeEnvelope returned multi-line JSON: %s", got)
	}
	if !strings.HasPrefix(string(got), `{"time":`) {
		t.Fatalf("time is not the first field: %s", got)
	}

	var decoded struct {
		Time          string          `json:"time"`
		Type          string          `json:"type"`
		BrowserUUID   string          `json:"browser_uuid"`
		SchemaVersion int             `json:"schema_version"`
		Params        json.RawMessage `json:"params"`
	}
	if err := json.Unmarshal(got, &decoded); err != nil {
		t.Fatalf("json.Unmarshal encoded envelope: %v", err)
	}

	gotTime, err := time.Parse(time.RFC3339Nano, decoded.Time)
	if err != nil {
		t.Fatalf("parse encoded time %q: %v", decoded.Time, err)
	}
	if !gotTime.Equal(envelope.Time) {
		t.Fatalf("time: got %v, want %v", gotTime, envelope.Time)
	}
	if !strings.HasSuffix(decoded.Time, "Z") {
		t.Fatalf("time missing Z suffix: %q", decoded.Time)
	}
	if decoded.Type != envelope.Type {
		t.Fatalf("type: got %q, want %q", decoded.Type, envelope.Type)
	}
	if decoded.BrowserUUID != envelope.BrowserUUID {
		t.Fatalf(
			"browser_uuid: got %q, want %q",
			decoded.BrowserUUID,
			envelope.BrowserUUID,
		)
	}
	if decoded.SchemaVersion != envelope.SchemaVersion {
		t.Fatalf(
			"schema_version: got %d, want %d",
			decoded.SchemaVersion,
			envelope.SchemaVersion,
		)
	}
	if string(decoded.Params) != string(envelope.Params) {
		t.Fatalf(
			"params: got %s, want %s",
			decoded.Params,
			envelope.Params,
		)
	}
}

func TestEncodeEnvelopeOmitsFractionalSecondsWhenWhole(t *testing.T) {
	t.Parallel()

	const wholeSecond = "2026-11-24T08:45:12Z"
	envelope := validEnvelope(t)
	envelope.Time = mustParseTime(t, wholeSecond)

	got, err := storage.EncodeEnvelope(envelope)
	if err != nil {
		t.Fatalf("EncodeEnvelope: %v", err)
	}
	wantPrefix := `{"time":"` + wholeSecond + `"`
	if !strings.HasPrefix(string(got), wantPrefix) {
		t.Fatalf("EncodeEnvelope: got %s, want prefix %s", got, wantPrefix)
	}
}

func TestEncodeEnvelopeKeepsEmptyParamsObject(t *testing.T) {
	t.Parallel()

	envelope := validEnvelope(t)
	envelope.Params = mustMarshal(t, map[string]any{})

	got, err := storage.EncodeEnvelope(envelope)
	if err != nil {
		t.Fatalf("EncodeEnvelope: %v", err)
	}

	var decoded struct {
		Params json.RawMessage `json:"params"`
	}
	if err := json.Unmarshal(got, &decoded); err != nil {
		t.Fatalf("json.Unmarshal encoded envelope: %v", err)
	}
	if string(decoded.Params) != string(envelope.Params) {
		t.Fatalf(
			"params: got %s, want %s",
			decoded.Params,
			envelope.Params,
		)
	}
}
