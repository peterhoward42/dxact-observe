package app_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/peterhoward42/dxact-observe/internal/app"
	"github.com/peterhoward42/dxact-observe/internal/event"
	"github.com/peterhoward42/dxact-observe/internal/storage"
)

// This file pins the Phase 2 POST / ingest contract for Application: content
// type, envelope validation, storage of the re-encoded envelope under the
// expected object name, the temporary AcceptedEventOut line, and the plan's
// status codes. Tests drive Application with a memory Store; they do not
// start the Functions Framework.

// TestPostAcceptsValidEnvelope checks the happy path end to end: 204, empty
// body, stored object name and compact JSON, and one AcceptedEventOut line.
func TestPostAcceptsValidEnvelope(t *testing.T) {
	t.Parallel()

	// Fixed event UUID so the expected object name is deterministic.
	eventID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	store := storage.NewMemoryStore(func() (uuid.UUID, error) {
		return eventID, nil
	})
	var accepted bytes.Buffer
	application := app.NewApplication(app.Dependencies{
		Store:            store,
		Frontend:         testFrontend(),
		AcceptedEventOut: &accepted,
	})
	body := mustMarshal(t, validPayload())

	rec := postJSON(t, application, "application/json", body)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusNoContent)
	}
	if got := readBody(t, rec); got != "" {
		t.Fatalf("body: got %q, want empty", got)
	}

	envelope, err := event.Parse(body)
	if err != nil {
		t.Fatalf("Parse fixture: %v", err)
	}
	wantName := storage.BuildObjectName(envelope.Time, eventID)
	names := store.ObjectNames()
	if len(names) != 1 || names[0] != wantName {
		t.Fatalf("ObjectNames: got %v, want [%s]", names, wantName)
	}
	stored, ok := store.GetBody(wantName)
	if !ok {
		t.Fatalf("GetBody(%q): missing", wantName)
	}
	want, err := storage.EncodeEnvelope(envelope)
	if err != nil {
		t.Fatalf("EncodeEnvelope: %v", err)
	}
	if string(stored) != string(want) {
		t.Fatalf("stored body: got %s, want %s", stored, want)
	}
	if got := accepted.String(); got != string(want)+"\n" {
		t.Fatalf("AcceptedEventOut: got %q, want %q", got, string(want)+"\n")
	}
}

// TestPostAcceptsJSONContentTypeWithCharset checks that an optional
// charset=utf-8 suffix on application/json is still accepted.
func TestPostAcceptsJSONContentTypeWithCharset(t *testing.T) {
	t.Parallel()

	store := storage.NewMemoryStore(uuid.NewRandom)
	application := newTestApp(t, store, nil)
	body := mustMarshal(t, validPayload())

	rec := postJSON(
		t,
		application,
		"application/json; charset=utf-8",
		body,
	)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusNoContent)
	}
	if got := len(store.ObjectNames()); got != 1 {
		t.Fatalf("stored objects: got %d, want 1", got)
	}
}

// TestPostRejectsWrongContentType checks that non-JSON media types return
// 415 and leave the store empty.
func TestPostRejectsWrongContentType(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		contentType string
	}{
		{name: "missing", contentType: ""},
		{name: "text plain", contentType: "text/plain"},
		{name: "json subtype", contentType: "application/jsonx"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := storage.NewMemoryStore(uuid.NewRandom)
			application := newTestApp(t, store, nil)
			body := mustMarshal(t, validPayload())

			rec := postJSON(t, application, tc.contentType, body)

			if rec.Code != http.StatusUnsupportedMediaType {
				t.Fatalf(
					"status: got %d, want %d",
					rec.Code,
					http.StatusUnsupportedMediaType,
				)
			}
			if got := len(store.ObjectNames()); got != 0 {
				t.Fatalf("stored objects: got %d, want 0", got)
			}
		})
	}
}

// TestPostRejectsInvalidEnvelope checks that bad JSON or a bad envelope
// returns 400 and does not store anything. Detailed envelope rules live in
// internal/event tests; here we only sample that Application maps them to 400.
func TestPostRejectsInvalidEnvelope(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		body []byte
	}{
		{
			name: "malformed JSON",
			body: []byte(`{"time":`),
		},
		{
			name: "missing type",
			body: mustMarshal(t, func() map[string]any {
				payload := validPayload()
				delete(payload, "type")
				return payload
			}()),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := storage.NewMemoryStore(uuid.NewRandom)
			application := newTestApp(t, store, nil)

			rec := postJSON(t, application, "application/json", tc.body)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf(
					"status: got %d, want %d",
					rec.Code,
					http.StatusBadRequest,
				)
			}
			if got := len(store.ObjectNames()); got != 0 {
				t.Fatalf("stored objects: got %d, want 0", got)
			}
		})
	}
}

// TestPostStorageFailureReturnsInternalServerError checks that a Store.Save
// error becomes HTTP 500.
func TestPostStorageFailureReturnsInternalServerError(t *testing.T) {
	t.Parallel()

	application := newTestApp(
		t,
		failingStore{err: errors.New("disk full")},
		nil,
	)
	body := mustMarshal(t, validPayload())

	rec := postJSON(t, application, "application/json", body)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"status: got %d, want %d",
			rec.Code,
			http.StatusInternalServerError,
		)
	}
}

// TestPostOutsideRootIsMethodNotAllowed checks that ingest is POST / only.
// A POST to another path is 405 with Allow listing GET and HEAD (not POST).
func TestPostOutsideRootIsMethodNotAllowed(t *testing.T) {
	t.Parallel()

	application := newTestApp(t, nil, nil)
	body := mustMarshal(t, validPayload())
	req := httptest.NewRequest(
		http.MethodPost,
		"/events",
		strings.NewReader(string(body)),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	application.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"status: got %d, want %d",
			rec.Code,
			http.StatusMethodNotAllowed,
		)
	}
	if got := rec.Header().Get("Allow"); got != "GET, HEAD" {
		t.Fatalf("Allow: got %q, want %q", got, "GET, HEAD")
	}
}
