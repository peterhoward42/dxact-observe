// Package storage provides storage for validated telemetry envelopes.
// The Store interface abstracts the underlying storage mechanism.
// Implementations encode envelopes as compact JSON objects.
//
// Phase 2 uses MemoryStore. A Google Cloud Storage (GCS) implementation
// will be introduced later behind the same interface.
package storage

import (
	"context"

	"github.com/peterhoward42/dxact-observe/internal/event"
)

// Store stores validated telemetry envelopes.
type Store interface {
	// Save stores a validated envelope under a new object name.
	// It must not overwrite an existing object.
	Save(ctx context.Context, envelope event.Envelope) error
}
