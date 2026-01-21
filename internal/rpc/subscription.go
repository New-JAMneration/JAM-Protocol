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
	subscriptions map[string]*Subscription // subID -> Subscription
	mu            sync.RWMutex
	done          chan struct{} // to signal shutdown all subscriptions
}

type Subscription struct {
	ID            string                // subscription ID
	EventType     eventbus.EventType    // type of events subscribed to
	EventBusSubID string                // event bus subscription ID
	EventCh       <-chan eventbus.Event // channel to receive events
	StopCh        chan struct{}         // channel to signal stopping the subscription
}

func NewSubscriptionManager(conn *websocket.Conn) *SubscriptionManager {
	return &SubscriptionManager{
		conn:          conn,
		subscriptions: make(map[string]*Subscription),
		done:          make(chan struct{}),
	}
}

func (sm *SubscriptionManager) Subscribe(eventType eventbus.EventType) (string, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	eventBusSubID, eventCh := eventbus.GetInstance().Subscribe(eventType, 50)
	subID := eventBusSubID
	sub := &Subscription{
		ID:            subID,
		EventType:     eventType,
		EventBusSubID: eventBusSubID,
		EventCh:       eventCh,
		StopCh:        make(chan struct{}),
	}

	sm.subscriptions[subID] = sub

	go sm.listenEvents(sub)

	logger.Info(fmt.Sprintf("Created subscription %s for event type %s", subID, eventType))

	return subID, nil
}

func (sm *SubscriptionManager) Unsubscribe(subID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sub, exists := sm.subscriptions[subID]
	if !exists {
		return fmt.Errorf("subscription %s not found", subID)
	}

	// Signal the subscription to stop
	close(sub.StopCh)

	// Unsubscribe from event bus
	eventbus.GetInstance().Unsubscribe(sub.EventType, sub.EventBusSubID)

	delete(sm.subscriptions, subID)

	logger.Info(fmt.Sprintf("Unsubscribed from subscription %s", subID))

	return nil
}

func (sm *SubscriptionManager) UnsubscribeAll() {
	sm.mu.Lock()

	logger.Info(fmt.Sprintf("Unsubscribing from all %d subscriptions", len(sm.subscriptions)))

	subs := make([]*Subscription, 0, len(sm.subscriptions))
	for _, sub := range sm.subscriptions {
		subs = append(subs, sub)
	}

	sm.subscriptions = make(map[string]*Subscription)
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

		eventbus.GetInstance().Unsubscribe(sub.EventType, sub.EventBusSubID)
	}
}

func (sm *SubscriptionManager) listenEvents(sub *Subscription) {
	for {
		select {
		case <-sub.StopCh:
			logger.Debug(fmt.Sprintf("Stopping event listener for subscription %s", sub.ID))
			return
		case <-sm.done:
			logger.Debug(fmt.Sprintf("Shutting down event listener for subscription %s", sub.ID))
			return
		case event, ok := <-sub.EventCh:
			if !ok {
				logger.Debug(fmt.Sprintf("Event channel closed for subscription %s", sub.ID))
				return
			}

			notification := sm.createNotification(sub.ID, event)

			if err := sm.sendNotification(notification); err != nil {
				logger.Error(fmt.Sprintf("Failed to send notification for subscription %s: %v", sub.ID, err))
			}
		}
	}
}

func (sm *SubscriptionManager) createNotification(subID string, event eventbus.Event) *JSONRPCNotification {
	return &JSONRPCNotification{
		JSONRPC: "2.0",
		Method:  subID,
		Params: NotificationParams{
			Subscription: subID,
			Result:       event.Data,
		},
	}
}

func (sm *SubscriptionManager) sendNotification(notification *JSONRPCNotification) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	logger.Debug(fmt.Sprintf("Sending notification: %s", string(data)))

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
