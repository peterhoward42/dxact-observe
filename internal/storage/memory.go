package storage

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/google/uuid"
	"github.com/peterhoward42/dxact-observe/internal/event"
)

// MemoryStore is an in-memory implementation of Store.
// Stored objects are lost when the process terminates.
type MemoryStore struct {
	mu      sync.Mutex
	objects map[string][]byte
	// newUUID provides UUIDs for object names.
	newUUID func() (uuid.UUID, error)
}

// NewMemoryStore returns an empty MemoryStore.
// newUUID must be non-nil; pass uuid.NewRandom for production.
func NewMemoryStore(newUUID func() (uuid.UUID, error)) *MemoryStore {
	if newUUID == nil {
		panic("storage: NewMemoryStore newUUID is nil")
	}
	return &MemoryStore{
		objects: make(map[string][]byte),
		newUUID: newUUID,
	}
}

// Save encodes envelope and stores it under a fresh object name from the
// envelope time and a new UUID. It does not overwrite an existing name.
func (m *MemoryStore) Save(
	ctx context.Context,
	envelope event.Envelope,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	id, err := m.newUUID()
	if err != nil {
		return fmt.Errorf("mint event UUID: %w", err)
	}
	name := BuildObjectName(envelope.Time, id)
	body, err := EncodeEnvelope(envelope)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.objects[name]; exists {
		return fmt.Errorf("object already exists: %s", name)
	}
	m.objects[name] = append([]byte(nil), body...)
	return nil
}

// GetBody returns a copy of the named object's body.
// If the object does not exist, it returns nil and false.
func (m *MemoryStore) GetBody(name string) ([]byte, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	body, ok := m.objects[name]
	if !ok {
		return nil, false
	}
	return append([]byte(nil), body...), true
}

// ObjectNames returns all stored object names in lexicographical order.
func (m *MemoryStore) ObjectNames() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	var names []string
	for name := range m.objects {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
