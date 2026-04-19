package events

import (
	"github.com/bbridges_11/document-registry/internal/events"
	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"go.uber.org/zap"
)

func LoggingMiddleware(logger *zap.Logger) outbound.Middleware {
	return func(next outbound.HandlerFunc) outbound.HandlerFunc {
		return func(envelope events.Envelope) {
			logger.Info("handling event",
				zap.String("event_id", envelope.Metadata.ID),
				zap.String("event_type", string(envelope.Event.Type())),
				zap.String("trace_id", envelope.Metadata.TraceID),
				zap.Time("occurred_at", envelope.Event.OccurredAt()),
			)
			next(envelope)
		}
	}
}

func RecoveryMiddleware(logger *zap.Logger) outbound.Middleware {
	return func(next outbound.HandlerFunc) outbound.HandlerFunc {
		return func(envelope events.Envelope) {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("panic in event handler",
						zap.String("event_type", string(envelope.Event.Type())),
						zap.Any("panic", r),
					)
				}
			}()
			next(envelope)
		}
	}
}
