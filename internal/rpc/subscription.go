package rpc

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/eventbus"
	"github.com/New-JAMneration/JAM-Protocol/logger"
	"github.com/gorilla/websocket"
)

type SubscriptionManager struct {
	conn          *websocket.Conn
	writeMu       *sync.Mutex              // shared write mutex with main handler
	eventSub      EventSubscriber          // injected event subscriber
	subscriptions map[uint64]*Subscription // subID -> Subscription
	mu            sync.RWMutex
	done          chan struct{} // to signal shutdown all subscriptions
}

type Subscription struct {
	ID            uint64                           // subscription ID (numeric per JIP-2)
	MethodName    string                           // original subscribe method name(e.g. "subscribeBestBlock")
	EventType     eventbus.EventType               // type of events subscribed to
	EventBusSubID uint64                           // event bus subscription ID
	EventCh       <-chan eventbus.Event            // channel to receive events
	StopCh        chan struct{}                    // channel to signal stopping the subscription
	Transform     func(eventbus.Event) interface{} // optional transform to reshape event data for notification
}

func NewSubscriptionManager(conn *websocket.Conn, writeMu *sync.Mutex, eventSub EventSubscriber) *SubscriptionManager {
	return &SubscriptionManager{
		conn:          conn,
		writeMu:       writeMu,
		eventSub:      eventSub,
		subscriptions: make(map[uint64]*Subscription),
		done:          make(chan struct{}),
	}
}

func (sm *SubscriptionManager) Subscribe(eventType eventbus.EventType, methodName string, transform ...func(eventbus.Event) interface{}) (uint64, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	eventBusSubID, eventCh := sm.eventSub.Subscribe(eventType, 50)
	subID := eventBusSubID
	sub := &Subscription{
		ID:            subID,
		MethodName:    methodName,
		EventType:     eventType,
		EventBusSubID: eventBusSubID,
		EventCh:       eventCh,
		StopCh:        make(chan struct{}),
	}
	if len(transform) > 0 {
		sub.Transform = transform[0]
	}

	sm.subscriptions[subID] = sub

	go sm.listenEvents(sub)

	logger.Info(fmt.Sprintf("Created subscription %d for event type %s", subID, eventType))

	return subID, nil
}

func (sm *SubscriptionManager) Unsubscribe(subID uint64) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sub, exists := sm.subscriptions[subID]
	if !exists {
		return fmt.Errorf("subscription %d not found", subID)
	}

	// Signal the subscription to stop
	close(sub.StopCh)

	// Unsubscribe from event bus
	sm.eventSub.Unsubscribe(sub.EventType, sub.EventBusSubID)

	delete(sm.subscriptions, subID)

	logger.Info(fmt.Sprintf("Unsubscribed from subscription %d", subID))

	return nil
}

func (sm *SubscriptionManager) UnsubscribeAll() {
	sm.mu.Lock()

	logger.Info(fmt.Sprintf("Unsubscribing from all %d subscriptions", len(sm.subscriptions)))

	subs := make([]*Subscription, 0, len(sm.subscriptions))
	for _, sub := range sm.subscriptions {
		subs = append(subs, sub)
	}

	sm.subscriptions = make(map[uint64]*Subscription)
	sm.mu.Unlock()

	select {
	case <-sm.done:
		// already closed
	default:
		close(sm.done)
	}

	for _, sub := range subs {
		select {
		case <-sub.StopCh:
			// already stopped
		default:
			close(sub.StopCh)
		}

		sm.eventSub.Unsubscribe(sub.EventType, sub.EventBusSubID)
	}
}

func (sm *SubscriptionManager) listenEvents(sub *Subscription) {
	for {
		select {
		case <-sub.StopCh:
			logger.Debug(fmt.Sprintf("Stopping event listener for subscription %d", sub.ID))
			return
		case <-sm.done:
			logger.Debug(fmt.Sprintf("Shutting down event listener for subscription %d", sub.ID))
			return
		case event, ok := <-sub.EventCh:
			if !ok {
				logger.Debug(fmt.Sprintf("Event channel closed for subscription %d", sub.ID))
				return
			}

			notification := sm.createNotification(sub, event)

			if err := sm.sendNotification(notification); err != nil {
				logger.Error(fmt.Sprintf("Failed to send notification for subscription %d: %v", sub.ID, err))
			}
		}
	}
}

func (sm *SubscriptionManager) createNotification(sub *Subscription, event eventbus.Event) *JSONRPCNotification {
	result := event.Data
	if sub.Transform != nil {
		result = sub.Transform(event)
	}

	return &JSONRPCNotification{
		JSONRPC: "2.0",
		Method:  sub.MethodName,
		Params: NotificationParams{
			Subscription: sub.ID,
			Result:       result,
		},
	}
}

func (sm *SubscriptionManager) sendNotification(notification *JSONRPCNotification) error {
	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	logger.Debug(fmt.Sprintf("Sending notification: %s", string(data)))

	sm.writeMu.Lock()
	defer sm.writeMu.Unlock()
	if err := sm.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("failed to write notification message: %w", err)
	}

	return nil
}

func (sm *SubscriptionManager) GetSubscriptionCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.subscriptions)
}
