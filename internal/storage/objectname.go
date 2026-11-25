package storage

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// BuildObjectName builds the storage object name for an event:
// events/YYYY-MM-DD/<event-uuid>.json.
//
// The date is derived from the event time in UTC, not the save time.
// The caller must supply a fresh per-event UUID, not the browser UUID.
func BuildObjectName(when time.Time, eventID uuid.UUID) string {
	date := when.UTC().Format("2006-01-02")
	return fmt.Sprintf("events/%s/%s.json", date, eventID.String())
}
