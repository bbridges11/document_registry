package events

import (
	"time"

	"github.com/google/uuid"
)

type Metadata struct {
	ID        string
	TraceID   string
	Source    string
	Timestamp time.Time
}

type Envelope struct {
	Event    Event
	Metadata Metadata
}

func NewEnvelope(event Event, traceID string) Envelope {
	return Envelope{
		Event: event,
		Metadata: Metadata{
			ID:        uuid.New().String(),
			TraceID:   traceID,
			Source:    "document-registry",
			Timestamp: time.Now().UTC(),
		},
	}
}
