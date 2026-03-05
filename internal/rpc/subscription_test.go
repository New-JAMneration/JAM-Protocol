package rpc

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/eventbus"
	"github.com/gorilla/websocket"
)

func TestSubscription(t *testing.T) {
	server := NewRPCServer(":19803")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := server.Start(ctx); err != nil && err != http.ErrServerClosed {
			t.Fatalf("Failed to start RPC server: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:19803/", nil)
	if err != nil {
		t.Fatalf("WebSocket dial error: %v", err)
	}

	notificationCh := make(chan *JSONRPCNotification, 10)
	responseCh := make(chan *JSONRPCResponse, 10)
	readerDone := make(chan struct{})
	go func() {
		defer close(notificationCh)
		defer close(responseCh)
		defer close(readerDone)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var rawMsg map[string]interface{}
			if err := json.Unmarshal(msg, &rawMsg); err != nil {
				t.Logf("Failed to unmarshal raw message: %v", err)
				continue
			}

			if _, hasID := rawMsg["id"]; hasID {
				var response JSONRPCResponse
				if err := json.Unmarshal(msg, &response); err != nil {
					t.Logf("Failed to unmarshal response: %v", err)
					continue
				}
				t.Logf("Received response: %s", string(msg))
				responseCh <- &response
				continue
			}
			var notification JSONRPCNotification
			if err := json.Unmarshal(msg, &notification); err != nil {
				t.Logf("Failed to unmarshal notification: %v", err)
				continue
			}
			t.Logf("Received notification: %+v", notification)
			notificationCh <- &notification
		}
	}()
	defer func() {
		conn.Close()
		<-readerDone
	}()

	subscribeReq := `{"jsonrpc":"2.0","id":1,"method":"subscribeBestBlock","params":[]}`
	err = conn.WriteMessage(websocket.TextMessage, []byte(subscribeReq))
	if err != nil {
		t.Fatalf("Failed to send subscribe request: %v", err)
	}

	var subID string
	select {
	case subscribeResp := <-responseCh:
		if subscribeResp == nil {
			t.Fatalf("Response channel closed unexpectedly")
		}

		if subscribeResp.Error != nil {
			t.Fatalf("Subscribe request returned error: %v", subscribeResp.Error)
		}

		var ok bool
		subID, ok = subscribeResp.Result.(string)
		if !ok {
			t.Fatalf("Invalid subscription ID in response: %v", subscribeResp.Result)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("Did not receive subscribe response in time")
	}

	testEvent := eventbus.Event{
		Type: eventbus.EventNewBlock,
		Data: eventbus.BlockEvent{
			HeaderHash: "test-hash-123",
			Slot:       456,
		},
	}
	eventbus.GetInstance().Publish(testEvent)

	select {
	case notification := <-notificationCh:
		if notification.Method != subID {
			t.Fatalf("Unexpected notification method: %s", notification.Method)
		}

		if notification.Params.Subscription != subID {
			t.Fatalf("Unexpected notification subscription ID: %s", notification.Params.Subscription)
		}

		resultMap, ok := notification.Params.Result.(map[string]interface{})
		if !ok {
			t.Fatalf("Invalid notification result format")
		}

		if resultMap["header_hash"] != "test-hash-123" {
			t.Errorf("Unexpected header_hash in notification: %v", resultMap["header_hash"])
		}

		if resultMap["slot"] != float64(456) {
			t.Errorf("Unexpected slot in notification: %v", resultMap["slot"])
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("Did not receive notification in time")
	}

	// Unsubscribe
	unsubscribeReq := `{"jsonrpc":"2.0","id":2,"method":"unsubscribe","params":["` + subID + `"]}`
	err = conn.WriteMessage(websocket.TextMessage, []byte(unsubscribeReq))
	if err != nil {
		t.Fatalf("Failed to send unsubscribe request: %v", err)
	}

	select {
	case unsubResp := <-responseCh:
		if unsubResp == nil {
			t.Fatalf("Response channel closed unexpectedly")
		}

		if unsubResp.Error != nil {
			t.Fatalf("Unsubscribe request returned error: %v", unsubResp.Error)
		}

		result, ok := unsubResp.Result.(bool)
		if !ok || !result {
			t.Fatalf("Unexpected unsubscribe result: %v", unsubResp.Result)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("Did not receive unsubscribe response in time")
	}

	eventbus.GetInstance().Publish(testEvent)

	select {
	case notification := <-notificationCh:
		t.Fatalf("Received unexpected notification after unsubscribe: %v", notification)
	case <-time.After(1 * time.Second):
		// No notification received, as expected
	}
}

func TestMultipleSubscriptions(t *testing.T) {
	server := NewRPCServer(":19804")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := server.Start(ctx); err != nil && err != http.ErrServerClosed {
			t.Fatalf("Failed to start RPC server: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:19804/", nil)
	if err != nil {
		t.Fatalf("WebSocket dial error: %v", err)
	}

	notificationCh := make(chan *JSONRPCNotification, 10)
	responseCh := make(chan *JSONRPCResponse, 10)
	readerDone := make(chan struct{})
	go func() {
		defer close(notificationCh)
		defer close(responseCh)
		defer close(readerDone)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var rawMsg map[string]interface{}
			if err := json.Unmarshal(msg, &rawMsg); err != nil {
				t.Logf("Failed to unmarshal raw message: %v", err)
				continue
			}

			if _, hasID := rawMsg["id"]; hasID {
				var response JSONRPCResponse
				if err := json.Unmarshal(msg, &response); err != nil {
					t.Logf("Failed to unmarshal response: %v", err)
					continue
				}
				t.Logf("Received response: %s", string(msg))
				responseCh <- &response
				continue
			}
			var notification JSONRPCNotification
			if err := json.Unmarshal(msg, &notification); err != nil {
				t.Logf("Failed to unmarshal notification: %v", err)
				continue
			}
			t.Logf("Received notification: %+v", notification)
			notificationCh <- &notification
		}
	}()
	defer func() {
		conn.Close()
		<-readerDone
	}()

	// Subscribe to best block notifications twice
	subscribeReq1 := `{"jsonrpc":"2.0","id":1,"method":"subscribeBestBlock","params":[]}`
	err = conn.WriteMessage(websocket.TextMessage, []byte(subscribeReq1))
	if err != nil {
		t.Fatalf("Failed to send first subscribe request: %v", err)
	}

	var subID1 string
	select {
	case resp1 := <-responseCh:
		if resp1.Error != nil {
			t.Fatalf("First subscribe request returned error: %v", resp1.Error)
		}
		subID1 = resp1.Result.(string)
		t.Logf("First subscription ID: %s", subID1)
	case <-time.After(2 * time.Second):
		t.Fatalf("Did not receive first subscribe response in time")
	}

	subscribeReq2 := `{"jsonrpc":"2.0","id":2,"method":"subscribeFinalizedBlock","params":[]}`
	err = conn.WriteMessage(websocket.TextMessage, []byte(subscribeReq2))
	if err != nil {
		t.Fatalf("Failed to send second subscribe request: %v", err)
	}

	var subID2 string
	select {
	case resp2 := <-responseCh:
		if resp2.Error != nil {
			t.Fatalf("Second subscribe request returned error: %v", resp2.Error)
		}
		subID2 = resp2.Result.(string)
		t.Logf("Second subscription ID: %s", subID2)
	case <-time.After(2 * time.Second):
		t.Fatalf("Did not receive second subscribe response in time")
	}

	// Publish events for both subscriptions
	eventbus.GetInstance().Publish(eventbus.Event{
		Type: eventbus.EventNewBlock,
		Data: eventbus.BlockEvent{
			HeaderHash: "new-block",
			Slot:       100,
		},
	})
	eventbus.GetInstance().Publish(eventbus.Event{
		Type: eventbus.EventFinalizedBlock,
		Data: eventbus.BlockEvent{
			HeaderHash: "finalized-block",
			Slot:       200,
		},
	})
	receivedSubs := make(map[string]bool)
	timeout := time.After(3 * time.Second)
	for len(receivedSubs) < 2 {
		select {
		case notification := <-notificationCh:
			subID := notification.Params.Subscription
			receivedSubs[subID] = true
			t.Logf("Received notification for subscription %s", subID)
		case <-timeout:
			t.Fatalf("Timeout: only received notifications for subscriptions: %v", receivedSubs)
		}
	}

	if !receivedSubs[subID1] {
		t.Fatalf("Did not receive notification for first subscription %s", subID1)
	}
	if !receivedSubs[subID2] {
		t.Fatalf("Did not receive notification for second subscription %s", subID2)
	}
}
