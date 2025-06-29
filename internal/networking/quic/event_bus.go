package quic

import (
	"context"
	"fmt"
	"sync"
)

type EventType string

const (
	BlockAuthored             EventType = "BlockAuthored"
	SafroleTicketsGenerated   EventType = "SafroleTicketsGenerated"
	WorkReportReceived        EventType = "WorkReportReceived"
	WorkPackagesReceived      EventType = "WorkPackagesReceived"
	WorkPackagesSubmitted     EventType = "WorkPackagesSubmitted"
	WorkPackageBundleReceived EventType = "WorkPackageBundleReceived"
	BlockImported             EventType = "BlockImported"
	BlockFinalized            EventType = "BlockFinalized"
	SafroleTicketsReceived    EventType = "SafroleTicketsReceived"
	WorkPackageBundleReady    EventType = "WorkPackageBundleReady"
	BeforeEpochChange         EventType = "BeforeEpochChange"
	WorkReportGenerated       EventType = "WorkReportGenerated"
	PeerAdded                 EventType = "PeerAdded"
	PeerUpdated               EventType = "PeerUpdated"
	BulkSyncCompleted         EventType = "BulkSyncCompleted"
)

type Event interface{}

type Handler func(ctx context.Context, event Event) error

type EventBus struct {
	sync.RWMutex
	handlers map[Event][]Handler
}

func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[Event][]Handler),
	}
}

func (eb *EventBus) Subscribe(eventType EventType, handler Handler) {
	eb.Lock()
	defer eb.Unlock()

	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

func (eb *EventBus) Unsubscribe(eventType EventType) {
	eb.Lock()
	defer eb.Unlock()

	if _, exists := eb.handlers[eventType]; exists {
		delete(eb.handlers, eventType)
	} else {
		fmt.Printf("No handlers found for event type: %s\n", eventType)
	}
}

func (eb *EventBus) Publish(ctx context.Context, event Event) error {
	eb.RLock()
	handlers := eb.handlers[fmt.Sprintf("%T", event)]
	eb.RUnlock()

	if len(handlers) == 0 {
		return nil
	}

	chain := func(ctx context.Context, event Event) error {
		for _, handler := range handlers {
			if err := handler(ctx, event); err != nil {
				return err
			}
		}
		return nil
	}

	return chain(ctx, event)
}

func (eb *EventBus) WaitFor(ctx context.Context, eventType EventType, timeout int) (Event, error) {
	eventCh := make(chan Event, 1)
	errCh := make(chan error, 1)

	handler := func(ctx context.Context, event Event) error {
		eventCh <- event
		return nil
	}

	eb.Subscribe(eventType, handler)
	defer eb.Unsubscribe(eventType)

	select {
	case event := <-eventCh:
		return event, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
