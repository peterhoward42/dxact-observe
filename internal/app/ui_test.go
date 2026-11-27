package app_test

import (
	"net/http"
	"testing"
	"testing/fstest"
)

// This file pins GET/HEAD UI serving for Application against an injected
// fstest.MapFS: real files, SPA fallback to index.html for non-asset paths,
// 404 for missing assets, and 405 Allow headers. No Vite build or staged
// static/ directory is required.

// TestGetServesIndex checks that GET / returns the injected index.html bytes.
func TestGetServesIndex(t *testing.T) {
	t.Parallel()

	frontend := testFrontend()
	application := newTestApp(t, nil, frontend)

	rec := doRequest(t, application, http.MethodGet, "/")

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}
	want := string(frontend["index.html"].Data)
	if got := readBody(t, rec); got != want {
		t.Fatalf("body: got %q, want %q", got, want)
	}
}

// TestGetServesRealAsset checks that an existing file under /assets is served
// as-is rather than replaced by the SPA fallback.
func TestGetServesRealAsset(t *testing.T) {
	t.Parallel()

	frontend := testFrontend()
	application := newTestApp(t, nil, frontend)

	rec := doRequest(t, application, http.MethodGet, "/assets/app.js")

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}
	want := string(frontend["assets/app.js"].Data)
	if got := readBody(t, rec); got != want {
		t.Fatalf("body: got %q, want %q", got, want)
	}
}

// TestGetFallsBackToIndexForSPARoute checks that a missing non-asset path
// returns index.html with 200 so client-side routes can load.
func TestGetFallsBackToIndexForSPARoute(t *testing.T) {
	t.Parallel()

	frontend := testFrontend()
	application := newTestApp(t, nil, frontend)

	rec := doRequest(t, application, http.MethodGet, "/reports/today")

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}
	want := string(frontend["index.html"].Data)
	if got := readBody(t, rec); got != want {
		t.Fatalf("body: got %q, want %q", got, want)
	}
}

// TestGetMissingAssetReturnsNotFound checks that missing asset-like paths
// return 404 instead of falling back to index.html.
func TestGetMissingAssetReturnsNotFound(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		path string
	}{
		{name: "missing js under assets", path: "/assets/missing.js"},
		{name: "missing root js", path: "/missing.js"},
		{name: "missing assets path", path: "/assets/no-such-chunk"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			application := newTestApp(t, nil, testFrontend())
			rec := doRequest(t, application, http.MethodGet, tc.path)

			if rec.Code != http.StatusNotFound {
				t.Fatalf(
					"status: got %d, want %d",
					rec.Code,
					http.StatusNotFound,
				)
			}
		})
	}
}

// TestHeadServesIndexWithoutBody checks that HEAD / succeeds like GET but
// carries an empty body.
func TestHeadServesIndexWithoutBody(t *testing.T) {
	t.Parallel()

	application := newTestApp(t, nil, testFrontend())

	rec := doRequest(t, application, http.MethodHead, "/")

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}
	if got := readBody(t, rec); got != "" {
		t.Fatalf("HEAD body: got %q, want empty", got)
	}
}

// TestUnsupportedMethodOnRoot checks that / allows GET, HEAD, and POST.
func TestUnsupportedMethodOnRoot(t *testing.T) {
	t.Parallel()

	application := newTestApp(t, nil, testFrontend())

	rec := doRequest(t, application, http.MethodPut, "/")

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"status: got %d, want %d",
			rec.Code,
			http.StatusMethodNotAllowed,
		)
	}
	if got := rec.Header().Get("Allow"); got != "GET, HEAD, POST" {
		t.Fatalf("Allow: got %q, want %q", got, "GET, HEAD, POST")
	}
}

// TestUnsupportedMethodOnAssetPath checks that non-root paths allow only
// GET and HEAD (ingest POST is not offered there).
func TestUnsupportedMethodOnAssetPath(t *testing.T) {
	t.Parallel()

	application := newTestApp(t, nil, testFrontend())

	rec := doRequest(t, application, http.MethodDelete, "/assets/app.js")

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

func TestGetRequiresIndexForRoot(t *testing.T) {
	t.Parallel()

	// Frontend without index.html: root cannot be served or fallen back.
	frontend := fstest.MapFS{
		"assets/app.js": &fstest.MapFile{Data: []byte("x")},
	}
	application := newTestApp(t, nil, frontend)

	rec := doRequest(t, application, http.MethodGet, "/")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusNotFound)
	}
}
