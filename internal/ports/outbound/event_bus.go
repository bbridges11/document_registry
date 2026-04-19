package outbound

import (
	"context"

	"github.com/bbridges_11/document-registry/internal/events"
)

type HandlerFunc func(events.Envelope)
type Middleware func(HandlerFunc) HandlerFunc

type EventBus interface {
	Publish(ctx context.Context, envelopes ...events.Envelope)
	Subscribe(eventType events.EventType, handler HandlerFunc)
	Start(ctx context.Context) error
	Stop() error
}
