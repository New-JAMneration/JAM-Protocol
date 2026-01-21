package eventbus

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/New-JAMneration/JAM-Protocol/logger"
)

type EventBus struct {
	mu          sync.RWMutex
	subscribers map[EventType]map[string]chan Event // eventType -> subID -> channel
	nextID      atomic.Uint64
}

var (
	instance *EventBus
	once     sync.Once
)

func GetInstance() *EventBus {
	once.Do(func() {
		instance = &EventBus{
			subscribers: make(map[EventType]map[string]chan Event),
		}
		logger.Info("EventBus Initialized")
	})
	return instance
}

func (eb *EventBus) Subscribe(eventType EventType, bufferSize int) (string, <-chan Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	id := fmt.Sprintf("sub-%d", eb.nextID.Add(1))
	ch := make(chan Event, bufferSize)

	if eb.subscribers[eventType] == nil {
		eb.subscribers[eventType] = make(map[string]chan Event)
	}

	eb.subscribers[eventType][id] = ch

	logger.Debug(fmt.Sprintf("New subscription: %s for event type %s (buffer: %d)", id, eventType, bufferSize))

	return id, ch
}

func (eb *EventBus) Unsubscribe(eventType EventType, subID string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if subscribers, ok := eb.subscribers[eventType]; ok {
		if ch, exists := subscribers[subID]; exists {
			close(ch)
			delete(subscribers, subID)
			logger.Debug(fmt.Sprintf("Unsubscribed: %s from event type %s", subID, eventType))

			// Clean up if no more subscribers for this event type
			if len(subscribers) == 0 {
				delete(eb.subscribers, eventType)
			}
		}
	}
}

func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	subscribers, ok := eb.subscribers[event.Type]
	if !ok || len(subscribers) == 0 {
		// No subscribers, return early
		return
	}

	logger.Debug(fmt.Sprintf("Publishing event: %s to %d subscribers", event.Type, len(subscribers)))

	droppedCount := 0
	for subID, ch := range subscribers {
		select {
		case ch <- event:
			// Event sent successfully
		default:
			// Channel is full, drop the event for this subscriber
			droppedCount++
			logger.Warn(fmt.Sprintf("Event dropped for subscriber %s due to full channel", subID))
		}
	}

	if droppedCount > 0 {
		logger.Warn(fmt.Sprintf("Event %s dropped for %d/%d subscribers", event.Type, droppedCount, len(subscribers)))
	}
}

func (eb *EventBus) GetSubscriberCount(eventType EventType) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if subscribers, ok := eb.subscribers[eventType]; ok {
		return len(subscribers)
	}
	return 0
}

func (eb *EventBus) GetTotalSubscribers() int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	total := 0
	for _, subscribers := range eb.subscribers {
		total += len(subscribers)
	}
	return total
}
