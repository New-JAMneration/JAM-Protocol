package rpc

import (
	"context"
	"testing"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/eventbus"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func setupTestChain() {
	blockchain.ResetInstance()
	chainState := blockchain.GetInstance()

	genesisBlock := types.Block{
		Header: types.Header{
			Slot:   0,
			Parent: types.HeaderHash{},
		},
	}
	genesisState := types.State{}

	chainState.GenerateGenesisBlock(genesisBlock)
	chainState.GenerateGenesisState(genesisState)
}
func TestChainWatcher_BestBlock(t *testing.T) {
	setupTestChain()
	chainState := blockchain.GetInstance()
	bus := eventbus.GetInstance()

	subID, eventCh := bus.Subscribe(eventbus.EventNewBlock, 10)
	defer bus.Unsubscribe(eventbus.EventNewBlock, subID)

	watcher := NewChainWatcher(chainState, bus)
	watcher.Start()
	defer watcher.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	select {
	case event := <-eventCh:
		blockEvent, ok := event.Data.(eventbus.BlockEvent)
		if !ok {
			t.Fatalf("Expected BlockEvent data type, got %T", event.Data)
		}
		t.Logf("Received new block event: Hash=%s, Slot=%d", blockEvent.HeaderHash, blockEvent.Slot)
		if blockEvent.HeaderHash == "" {
			t.Fatalf("Received block event with empty hash")
		}
	case <-ctx.Done():
		t.Fatalf("Did not receive new block event in time")
	}
}

func TestChainWatcher_FinalizedBlock(t *testing.T) {
	setupTestChain()
	chainState := blockchain.GetInstance()
	bus := eventbus.GetInstance()

	subID, eventCh := bus.Subscribe(eventbus.EventFinalizedBlock, 10)
	defer bus.Unsubscribe(eventbus.EventFinalizedBlock, subID)

	watcher := NewChainWatcher(chainState, bus)
	watcher.Start()
	defer watcher.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	select {
	case event := <-eventCh:
		blockEvent, ok := event.Data.(eventbus.BlockEvent)
		if !ok {
			t.Fatalf("Expected BlockEvent data type, got %T", event.Data)
		}
		t.Logf("Received finalized block event: Hash=%s, Slot=%d", blockEvent.HeaderHash, blockEvent.Slot)
		if blockEvent.HeaderHash == "" {
			t.Fatalf("Received finalized block event with empty hash")
		}
	case <-ctx.Done():
		t.Fatalf("Did not receive finalized block event in time")
	}
}

func TestChainWatcher_SyncState(t *testing.T) {
	setupTestChain()
	chainState := blockchain.GetInstance()
	bus := eventbus.GetInstance()

	subID, eventCh := bus.Subscribe(eventbus.EventSyncStateChanged, 10)
	defer bus.Unsubscribe(eventbus.EventSyncStateChanged, subID)

	watcher := NewChainWatcher(chainState, bus)
	watcher.Start()
	defer watcher.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	select {
	case event := <-eventCh:
		syncEvent, ok := event.Data.(eventbus.SyncStateEvent)
		if !ok {
			t.Fatalf("Expected SyncStateEvent data type, got %T", event.Data)
		}
		t.Logf("Received sync state change event: NewState=%s", syncEvent.Status)
		if syncEvent.Status == "" {
			t.Fatalf("Received sync state change event with empty state")
		}
	case <-ctx.Done():
		t.Fatalf("Did not receive sync state change event in time")
	}
}
