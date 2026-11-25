package storage_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/peterhoward42/dxact-observe/internal/storage"
)

func TestMemoryStoreSaveStoresEncodedEnvelopeUnderObjectName(t *testing.T) {
	t.Parallel()

	eventID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	store := storage.NewMemoryStore(fixedIDs(t, eventID))
	envelope := validEnvelope(t)
	wantName := storage.BuildObjectName(envelope.Time, eventID)
	wantBody, err := storage.EncodeEnvelope(envelope)
	if err != nil {
		t.Fatalf("EncodeEnvelope: %v", err)
	}

	if err := store.Save(context.Background(), envelope); err != nil {
		t.Fatalf("Save: %v", err)
	}

	gotBody, ok := store.GetBody(wantName)
	if !ok {
		t.Fatalf("GetBody(%q): missing", wantName)
	}
	if string(gotBody) != string(wantBody) {
		t.Fatalf("stored body: got %s, want %s", gotBody, wantBody)
	}
	names := store.ObjectNames()
	if len(names) != 1 || names[0] != wantName {
		t.Fatalf("ObjectNames: got %v, want [%s]", names, wantName)
	}
}

func TestMemoryStoreSaveKeepsDistinctNamesForSameBrowser(t *testing.T) {
	t.Parallel()

	id1 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	id2 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	store := storage.NewMemoryStore(fixedIDs(t, id1, id2))
	envelope := validEnvelope(t)

	if err := store.Save(context.Background(), envelope); err != nil {
		t.Fatalf("Save first: %v", err)
	}
	if err := store.Save(context.Background(), envelope); err != nil {
		t.Fatalf("Save second: %v", err)
	}

	want := []string{
		storage.BuildObjectName(envelope.Time, id1),
		storage.BuildObjectName(envelope.Time, id2),
	}
	got := store.ObjectNames()
	if !sameNames(got, want) {
		t.Fatalf("ObjectNames: got %v, want %v", got, want)
	}
	for _, name := range want {
		if _, ok := store.GetBody(name); !ok {
			t.Fatalf("GetBody(%q): missing", name)
		}
	}
}

func TestMemoryStoreSaveRejectsDuplicateObjectName(t *testing.T) {
	t.Parallel()

	eventID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	store := storage.NewMemoryStore(fixedIDs(t, eventID, eventID))
	envelope := validEnvelope(t)

	if err := store.Save(context.Background(), envelope); err != nil {
		t.Fatalf("Save first: %v", err)
	}
	err := store.Save(context.Background(), envelope)
	if err == nil {
		t.Fatal("Save second: accepted duplicate object name")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("error %q does not mention already exists", err.Error())
	}
	if got := store.ObjectNames(); len(got) != 1 {
		t.Fatalf("ObjectNames after duplicate: got %v, want one entry", got)
	}
}

func TestMemoryStoreSaveHonoursCancelledContext(t *testing.T) {
	t.Parallel()

	store := storage.NewMemoryStore(fixedIDs(
		t,
		uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
	))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := store.Save(ctx, validEnvelope(t))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Save cancelled context: got %v, want context.Canceled", err)
	}
	if got := store.ObjectNames(); len(got) != 0 {
		t.Fatalf("ObjectNames after cancel: got %v, want empty", got)
	}
}

func TestMemoryStoreSavePropagatesIDMintFailure(t *testing.T) {
	t.Parallel()

	mintErr := errors.New("mint failed")
	store := storage.NewMemoryStore(func() (uuid.UUID, error) {
		return uuid.Nil, mintErr
	})

	err := store.Save(context.Background(), validEnvelope(t))
	if !errors.Is(err, mintErr) {
		t.Fatalf("Save: got %v, want mint failure", err)
	}
	if got := store.ObjectNames(); len(got) != 0 {
		t.Fatalf("ObjectNames after mint failure: got %v, want empty", got)
	}
}

func TestMemoryStoreSaveIsSafeUnderConcurrentSaves(t *testing.T) {
	t.Parallel()

	const workers = 32
	ids := make([]uuid.UUID, workers)
	for i := range ids {
		ids[i] = uuid.Must(uuid.NewRandom())
	}
	store := storage.NewMemoryStore(fixedIDs(t, ids...))
	envelope := validEnvelope(t)
	wantBody, err := storage.EncodeEnvelope(envelope)
	if err != nil {
		t.Fatalf("EncodeEnvelope: %v", err)
	}
	wantNames := make([]string, workers)
	for i, id := range ids {
		wantNames[i] = storage.BuildObjectName(envelope.Time, id)
	}

	var wg sync.WaitGroup
	errs := make([]error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = store.Save(context.Background(), envelope)
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("Save worker %d: %v", i, err)
		}
	}
	gotNames := store.ObjectNames()
	if !sameNames(gotNames, wantNames) {
		t.Fatalf("ObjectNames: got %v, want %v", gotNames, wantNames)
	}
	for _, name := range wantNames {
		gotBody, ok := store.GetBody(name)
		if !ok {
			t.Fatalf("GetBody(%q): missing", name)
		}
		if string(gotBody) != string(wantBody) {
			t.Fatalf("stored body for %q: got %s, want %s", name, gotBody, wantBody)
		}
	}
}

// fixedIDs returns a newUUID func that yields ids in order, then fails the test.
// It is safe for concurrent callers.
func fixedIDs(
	t *testing.T,
	ids ...uuid.UUID,
) func() (uuid.UUID, error) {
	t.Helper()
	var mu sync.Mutex
	next := 0
	return func() (uuid.UUID, error) {
		mu.Lock()
		defer mu.Unlock()
		if next >= len(ids) {
			t.Fatalf("newUUID called more times than ids supplied")
		}
		id := ids[next]
		next++
		return id, nil
	}
}

func sameNames(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	counts := make(map[string]int, len(want))
	for _, name := range want {
		counts[name]++
	}
	for _, name := range got {
		counts[name]--
		if counts[name] < 0 {
			return false
		}
	}
	for _, n := range counts {
		if n != 0 {
			return false
		}
	}
	return true
}
