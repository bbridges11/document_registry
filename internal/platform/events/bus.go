package events

import (
	"context"
	"sync"

	"github.com/bbridges_11/document-registry/internal/events"
	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"go.uber.org/zap"
)

type Bus struct {
	logger        *zap.Logger
	subscriptions map[events.EventType][]outbound.HandlerFunc
	queue         chan events.Envelope
	middleware    []outbound.Middleware
	mu            sync.RWMutex
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	bufferSize    int
	workers       int
}

func NewBus(logger *zap.Logger, bufferSize, workers int) *Bus {
	ctx, cancel := context.WithCancel(context.Background())
	return &Bus{
		logger:        logger,
		subscriptions: make(map[events.EventType][]outbound.HandlerFunc),
		queue:         make(chan events.Envelope, bufferSize),
		middleware:    make([]outbound.Middleware, 0),
		ctx:           ctx,
		cancel:        cancel,
		bufferSize:    bufferSize,
		workers:       workers,
	}
}

func (b *Bus) Use(middleware outbound.Middleware) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.middleware = append(b.middleware, middleware)
}

func (b *Bus) Subscribe(eventType events.EventType, handler outbound.HandlerFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Wrap handler with middleware
	wrappedHandler := b.applyMiddleware(handler)
	b.subscriptions[eventType] = append(b.subscriptions[eventType], wrappedHandler)
}

func (b *Bus) Publish(ctx context.Context, envelopes ...events.Envelope) {
	for _, envelope := range envelopes {
		select {
		case b.queue <- envelope:
		case <-ctx.Done():
			b.logger.Warn("publish cancelled", zap.String("event_type", string(envelope.Event.Type())))
			return
		default:
			b.logger.Error("event queue full, dropping event", zap.String("event_type", string(envelope.Event.Type())))
		}
	}
}

func (b *Bus) Start(ctx context.Context) error {
	b.logger.Info("starting event bus", zap.Int("workers", b.workers), zap.Int("buffer_size", b.bufferSize))

	for i := 0; i < b.workers; i++ {
		b.wg.Add(1)
		go b.worker(i)
	}

	return nil
}

func (b *Bus) Stop() error {
	b.logger.Info("stopping event bus")
	b.cancel()
	close(b.queue)
	b.wg.Wait()
	b.logger.Info("event bus stopped")
	return nil
}

func (b *Bus) worker(id int) {
	defer b.wg.Done()

	for {
		select {
		case <-b.ctx.Done():
			return
		case envelope, ok := <-b.queue:
			if !ok {
				return
			}
			b.dispatch(envelope)
		}
	}
}

func (b *Bus) dispatch(envelope events.Envelope) {
	b.mu.RLock()
	handlers := b.subscriptions[envelope.Event.Type()]
	b.mu.RUnlock()

	for _, handler := range handlers {
		// Call handler in separate goroutine for non-blocking dispatch
		go func(h outbound.HandlerFunc, env events.Envelope) {
			h(env)
		}(handler, envelope)
	}
}

func (b *Bus) applyMiddleware(handler outbound.HandlerFunc) outbound.HandlerFunc {
	result := handler
	for i := len(b.middleware) - 1; i >= 0; i-- {
		result = b.middleware[i](result)
	}
	return result
}
