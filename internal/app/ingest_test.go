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

func TestPostAcceptsValidEnvelope(t *testing.T) {
	t.Parallel()

	store := storage.NewMemoryStore(uuid.NewRandom)
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

	names := store.ObjectNames()
	if len(names) != 1 {
		t.Fatalf("stored objects: got %d, want 1", len(names))
	}
	stored, ok := store.GetBody(names[0])
	if !ok {
		t.Fatalf("GetBody(%q): missing", names[0])
	}
	envelope, err := event.Parse(body)
	if err != nil {
		t.Fatalf("Parse fixture: %v", err)
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
