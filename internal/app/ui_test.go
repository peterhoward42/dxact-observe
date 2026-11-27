package app_test

import (
	"net/http"
	"testing"
	"testing/fstest"
)

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
