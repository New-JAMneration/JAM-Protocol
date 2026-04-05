package eventbus

import (
	"sync"
	"testing"
	"time"
)

func TestEventBusSingleton(t *testing.T) {
	eb1 := GetInstance()
	eb2 := GetInstance()

	if eb1 != eb2 {
		t.Fatalf("Expected singleton EventBus instances to be the same")
	}
}

func TestSubscribeAndPublish(t *testing.T) {
	eb := GetInstance()

	subID, eventCh := eb.Subscribe(EventNewBlock, 10)

	count := eb.GetSubscriberCount(EventNewBlock)
	if count != 1 {
		t.Fatalf("Expected 1 subscriber, got %d", count)
	}

	testEvent := Event{
		Type: EventNewBlock,
		Data: BlockEvent{
			HeaderHash: "test-hash",
			Slot:       123,
		},
	}

	eb.Publish(testEvent)

	select {
	case receivedEvent := <-eventCh:
		if receivedEvent.Type != EventNewBlock {
			t.Fatalf("Expected event type %s, got %s", EventNewBlock, receivedEvent.Type)
		}

		blockData, ok := receivedEvent.Data.(BlockEvent)
		if !ok {
			t.Fatalf("Expected BlockEvent data type")
		}

		if blockData.HeaderHash != "test-hash" || blockData.Slot != 123 {
			t.Fatalf("Received event data does not match published data")
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("Did not receive published event in time")
	}
	eb.Unsubscribe(EventNewBlock, subID)
}

func TestMultipleSubscribers(t *testing.T) {
	eb := GetInstance()

	subID1, eventCh1 := eb.Subscribe(EventFinalizedBlock, 5)
	subID2, eventCh2 := eb.Subscribe(EventFinalizedBlock, 5)
	subID3, eventCh3 := eb.Subscribe(EventFinalizedBlock, 5)

	count := eb.GetSubscriberCount(EventFinalizedBlock)
	if count != 3 {
		t.Fatalf("Expected 3 subscribers, got %d", count)
	}

	testEvent := Event{
		Type: EventFinalizedBlock,
		Data: BlockEvent{
			HeaderHash: "finalized-hash",
			Slot:       456,
		},
	}

	eb.Publish(testEvent)

	var wg sync.WaitGroup
	wg.Add(3)

	checkReveiver := func(ch <-chan Event, name string) {
		defer wg.Done()
		select {
		case event := <-ch:
			blockData := event.Data.(BlockEvent)
			if blockData.Slot != 456 {
				t.Errorf("%s: Received event data does not match published data", name)
			}
		case <-time.After(1 * time.Second):
			t.Errorf("%s: timeout", name)
		}
	}

	go checkReveiver(eventCh1, "Subscriber 1")
	go checkReveiver(eventCh2, "Subscriber 2")
	go checkReveiver(eventCh3, "Subscriber 3")

	wg.Wait()

	eb.Unsubscribe(EventFinalizedBlock, subID1)
	eb.Unsubscribe(EventFinalizedBlock, subID2)
	eb.Unsubscribe(EventFinalizedBlock, subID3)
}

func TestUnsubscribe(t *testing.T) {
	eb := GetInstance()

	subID, eventCh := eb.Subscribe(EventImportedBlock, 5)

	if eb.GetSubscriberCount(EventImportedBlock) != 1 {
		t.Error("Expected 1 subscriber")
	}

	eb.Unsubscribe(EventImportedBlock, subID)

	if eb.GetSubscriberCount(EventImportedBlock) != 0 {
		t.Error("Expected 0 subscribers after unsubscribe")
	}

	select {
	case _, ok := <-eventCh:
		if ok {
			t.Error("Expected channel to be closed after unsubscribe")
		}
	case <-time.After(500 * time.Millisecond):
		// Channel is still open
		t.Error("Expected channel to be closed after unsubscribe")
	}
}

func TestNonBlockingPublish(t *testing.T) {
	eb := GetInstance()

	subID, eventCh := eb.Subscribe(EventNewBlock, 2)

	for i := 0; i < 3; i++ {
		eb.Publish(Event{
			Type: EventNewBlock,
			Data: BlockEvent{Slot: uint64(i)},
		})
	}

	eb.Unsubscribe(EventNewBlock, subID)

	for len(eventCh) > 0 {
		<-eventCh
	}
}

func TestDifferentEventTypes(t *testing.T) {
	eb := GetInstance()

	subID1, eventCh1 := eb.Subscribe(EventNewBlock, 5)
	subID2, eventCh2 := eb.Subscribe(EventFinalizedBlock, 5)

	eb.Publish(Event{
		Type: EventNewBlock,
		Data: BlockEvent{Slot: 100},
	})

	select {
	case event := <-eventCh1:
		if event.Type != EventNewBlock {
			t.Errorf("Subscriber 1: Expected EventNewBlock, got %s", event.Type)
		}
	case <-time.After(1 * time.Second):
		t.Error("Subscriber 1: Did not receive EventNewBlock")
	}

	select {
	case <-eventCh2:
		t.Error("Subscriber 2: Should not have received EventNewBlock")
	case <-time.After(500 * time.Millisecond):
		// Expected timeout
	}

	eb.Publish(Event{
		Type: EventFinalizedBlock,
		Data: BlockEvent{Slot: 200},
	})

	select {
	case event := <-eventCh2:
		if event.Type != EventFinalizedBlock {
			t.Errorf("Subscriber 2: Expected EventFinalizedBlock, got %s", event.Type)
		}
	case <-time.After(1 * time.Second):
		t.Error("Subscriber 2: Did not receive EventFinalizedBlock")
	}

	select {
	case <-eventCh1:
		t.Error("Subscriber 1: Should not have received EventFinalizedBlock")
	case <-time.After(500 * time.Millisecond):
		// Expected timeout
	}

	eb.Unsubscribe(EventNewBlock, subID1)
	eb.Unsubscribe(EventFinalizedBlock, subID2)
}

func TestConcurrency(t *testing.T) {
	eb := GetInstance()

	const numSubscribers = 10
	const numEvents = 100

	var setupWg sync.WaitGroup
	setupWg.Add(numSubscribers)

	var receiveWg sync.WaitGroup
	receiveWg.Add(numSubscribers)

	for i := 0; i < numSubscribers; i++ {
		go func(id int) {
			defer receiveWg.Done()

			subID, eventCh := eb.Subscribe(EventNewBlock, 150)
			defer eb.Unsubscribe(EventNewBlock, subID)

			setupWg.Done()

			received := 0
			timeout := time.After(5 * time.Second)

			for received < numEvents {
				select {
				case <-eventCh:
					received++
				case <-timeout:
					t.Errorf("Subscriber %d: Timeout waiting for events", id)
					return
				}
			}
		}(i)
	}

	setupWg.Wait()
	time.Sleep(100 * time.Millisecond)

	for i := 0; i < numEvents; i++ {
		eb.Publish(Event{
			Type: EventNewBlock,
			Data: BlockEvent{Slot: uint64(i)},
		})
	}

	receiveWg.Wait()
}

func TestGetTotalSubscribers(t *testing.T) {
	eb := GetInstance()

	subID1, _ := eb.Subscribe(EventNewBlock, 5)
	subID2, _ := eb.Subscribe(EventNewBlock, 5)
	subID3, _ := eb.Subscribe(EventFinalizedBlock, 5)

	total := eb.GetTotalSubscribers()
	if total != 3 {
		t.Fatalf("Expected total subscribers to be 3, got %d", total)
	}

	eb.Unsubscribe(EventNewBlock, subID1)
	eb.Unsubscribe(EventNewBlock, subID2)
	eb.Unsubscribe(EventFinalizedBlock, subID3)
}
