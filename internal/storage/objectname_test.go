package storage_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/peterhoward42/dxact-observe/internal/storage"
)

func TestBuildObjectNameUsesUTCDateAndEventUUID(t *testing.T) {
	t.Parallel()

	when := mustParseTime(t, "2026-11-24T08:45:12.345Z")
	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	got := storage.BuildObjectName(when, id)
	want := "events/2026-11-24/550e8400-e29b-41d4-a716-446655440000.json"
	if got != want {
		t.Fatalf("BuildObjectName: got %q, want %q", got, want)
	}
}

func TestBuildObjectNameUsesUTCCalendarDayNotLocal(t *testing.T) {
	t.Parallel()

	// 2026-11-24 23:30 UTC+05:30 is still 2026-11-24 UTC.
	loc := time.FixedZone("UTC+5:30", 5*3600+30*60)
	when := time.Date(2026, 11, 25, 0, 30, 0, 0, loc)
	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	got := storage.BuildObjectName(when, id)
	want := "events/2026-11-24/550e8400-e29b-41d4-a716-446655440000.json"
	if got != want {
		t.Fatalf("BuildObjectName: got %q, want %q", got, want)
	}
}
