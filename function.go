// Package function is the Cloud Function composition root.
// It wires runtime dependencies, embeds the staged frontend, and
// registers one HTTP handler with the Functions Framework.
// Request handling lives in internal/app.
package function

import (
	"embed"
	"io/fs"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/google/uuid"

	"github.com/peterhoward42/dxact-observe/internal/app"
	"github.com/peterhoward42/dxact-observe/internal/storage"
)

// staticFS contains the built frontend files, copied from ui/dist
// to static/ by the Makefile before compilation.
//
//go:embed static
var staticFS embed.FS

// init assembles the application with its runtime dependencies
// and registers it as the Observe HTTP Cloud Function.
func init() {
	frontend, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic("function: embed static: " + err.Error())
	}

	application := app.NewApplication(app.Dependencies{
		Store:            storage.NewMemoryStore(uuid.NewRandom),
		Frontend:         frontend,
		AcceptedEventOut: os.Stdout,
	})

	functions.HTTP("Observe", application.ServeHTTP)
}
