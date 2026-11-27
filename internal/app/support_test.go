package app_test

import (
	"context"
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/google/uuid"
	"github.com/peterhoward42/dxact-observe/internal/app"
	"github.com/peterhoward42/dxact-observe/internal/event"
	"github.com/peterhoward42/dxact-observe/internal/storage"
)

// Shared helpers for Application HTTP tests in this package.

// failingStore is a Store that always returns err from Save.
type failingStore struct {
	err error
}

func (s failingStore) Save(
	ctx context.Context,
	envelope event.Envelope,
) error {
	return s.err
}

// newTestApp builds an Application for tests.
// A nil store or frontend is replaced with a memory store or testFrontend().
// AcceptedEventOut is left nil so ordinary tests do not assert the temporary
// stdout scaffolding; tests that care pass Dependencies themselves.
func newTestApp(
	t *testing.T,
	store storage.Store,
	frontend fs.FS,
) *app.Application {
	t.Helper()
	if store == nil {
		store = storage.NewMemoryStore(uuid.NewRandom)
	}
	if frontend == nil {
		frontend = testFrontend()
	}
	return app.NewApplication(app.Dependencies{
		Store:    store,
		Frontend: frontend,
	})
}

// testFrontend is a minimal built-UI tree for FileServer tests.
func testFrontend() fstest.MapFS {
	return fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte("<!doctype html><title>Hello</title>"),
		},
		"assets/app.js": &fstest.MapFile{
			Data: []byte("console.log('hi')"),
		},
	}
}

// validPayload returns a new valid envelope map for one test to mutate.
func validPayload() map[string]any {
	return map[string]any{
		"time":           "2026-11-24T08:45:12.345Z",
		"type":           "Application opened",
		"browser_uuid":   "8e6a4c1d-5f29-4e83-9e10-42bd735af0c2",
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

// postJSON issues POST / with the given Content-Type and body.
// An empty contentType omits the header so missing-type cases can be tested.
func postJSON(
	t *testing.T,
	handler http.Handler,
	contentType string,
	body []byte,
) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader(string(body)),
	)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

// doRequest issues a body-less HTTP request of the given method and path.
func doRequest(
	t *testing.T,
	handler http.Handler,
	method, path string,
) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func readBody(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	body, err := io.ReadAll(rec.Result().Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(body)
}
