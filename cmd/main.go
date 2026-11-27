// Command cmd starts the Observe Cloud Function locally via the
// Functions Framework. Production deploy uses the buildpack's own
// main; this is only for local use (make run).
package main

import (
	"log"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"

	// Blank-import so the root package registers the Observe handler.
	_ "github.com/peterhoward42/dxact-observe"
)

// main starts the local HTTP server, configured via PORT and LOCAL_ONLY.
// Set LOCAL_ONLY=true to prevent access from other devices on the network.
func main() {
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	hostname := ""
	if os.Getenv("LOCAL_ONLY") == "true" {
		hostname = "127.0.0.1"
	}

	addr := hostname + ":" + port
	if hostname == "" {
		addr = ":" + port
	}
	log.Printf("Observe listening on http://%s/", addr)

	if err := funcframework.StartHostPort(hostname, port); err != nil {
		log.Fatalf("funcframework.StartHostPort: %v\n", err)
	}
}
