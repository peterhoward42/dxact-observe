package app

import (
	"fmt"
	"io"
	"mime"
	"net/http"

	"github.com/peterhoward42/dxact-observe/internal/event"
	"github.com/peterhoward42/dxact-observe/internal/storage"
)

// handlePost receives one telemetry envelope at POST /.
func (a *Application) handlePost(w http.ResponseWriter, r *http.Request) {
	if cleanURLPath(r.URL.Path) != "/" {
		a.methodNotAllowed(w, r)
		return
	}
	if !isJSONContentType(r.Header.Get("Content-Type")) {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	envelope, err := event.Parse(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := a.deps.Store.Save(r.Context(), envelope); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Temporary local scaffolding: print accepted events so a manual POST
	// against a local run can be inspected in the terminal. Remove before
	// the first cloud deploy.
	if a.deps.AcceptedEventOut != nil {
		if encoded, err := storage.EncodeEnvelope(envelope); err == nil {
			fmt.Fprintln(a.deps.AcceptedEventOut, string(encoded))
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// isJSONContentType reports whether header is application/json, optionally
// with parameters such as charset=utf-8.
func isJSONContentType(header string) bool {
	mediaType, _, err := mime.ParseMediaType(header)
	if err != nil {
		return false
	}
	return mediaType == "application/json"
}
