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
