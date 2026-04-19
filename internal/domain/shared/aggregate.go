package shared

import "github.com/bbridges_11/document-registry/internal/events"

type EventEmitter interface {
	Events() []events.Event
	ClearEvents()
}

type AggregateRoot struct {
	events []events.Event
}

func (a *AggregateRoot) AddEvent(event events.Event) {
	a.events = append(a.events, event)
}

func (a *AggregateRoot) Events() []events.Event {
	return a.events
}

func (a *AggregateRoot) ClearEvents() {
	a.events = nil
}
