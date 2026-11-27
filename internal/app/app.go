// Package app handles incoming HTTP requests for the telemetry service.
// It receives telemetry events and serves the web interface.
// The Cloud Function creates an Application and passes incoming requests
// to it. Event validation and storage are handled by separate packages.
package app

import (
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/peterhoward42/dxact-observe/internal/storage"
)

// Dependencies holds the external resources that Application needs to
// handle HTTP requests: event storage, web interface files, and a
// temporary accepted-event writer.
type Dependencies struct {
	// Store saves telemetry envelopes that have passed validation.
	Store storage.Store
	// Frontend is the root of built UI files (index.html and assets).
	Frontend fs.FS
	// AcceptedEventOut writes each accepted event as a single JSON line.
	// Temporary local scaffolding for inspecting manual POSTs; the in-memory
	// store is not otherwise browsable. Production uses os.Stdout; nil
	// disables output (typically in tests).
	AcceptedEventOut io.Writer
}

// Application owns HTTP request handling for ingest and UI serving.
type Application struct {
	deps       Dependencies
	fileServer http.Handler
}

// NewApplication creates an Application using the supplied dependencies.
// Panics if Store or Frontend is nil.
func NewApplication(deps Dependencies) *Application {
	if deps.Store == nil {
		panic("app: NewApplication Store is nil")
	}
	if deps.Frontend == nil {
		panic("app: NewApplication Frontend is nil")
	}
	return &Application{
		deps:       deps,
		fileServer: http.FileServer(http.FS(deps.Frontend)),
	}
}

// ServeHTTP routes HTTP requests by method: POST ingests telemetry events,
// GET and HEAD serve the UI, and other methods return 405 Method Not Allowed.
func (a *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		a.handlePost(w, r)
	case http.MethodGet, http.MethodHead:
		a.serveFrontend(w, r)
	default:
		a.methodNotAllowed(w, r)
	}
}

// methodNotAllowed returns a 405 error and lists the HTTP methods
// permitted for the requested URL in the Allow header.
func (a *Application) methodNotAllowed(
	w http.ResponseWriter,
	r *http.Request,
) {
	if cleanURLPath(r.URL.Path) == "/" {
		w.Header().Set("Allow", "GET, HEAD, POST")
	} else {
		w.Header().Set("Allow", "GET, HEAD")
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

// cleanURLPath returns a cleaned absolute URL path.
func cleanURLPath(raw string) string {
	return path.Clean("/" + strings.TrimPrefix(raw, "/"))
}
