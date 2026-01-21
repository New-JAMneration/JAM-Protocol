package rpc

import (
	"context"
	"fmt"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/eventbus"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/New-JAMneration/JAM-Protocol/logger"
)

type ChainWatcher struct {
	chainState *blockchain.ChainState
	eventbus   *eventbus.EventBus

	latestBestBlockHash      types.HeaderHash
	latestFinalizedBlockHash types.HeaderHash
	latestSyncState          string

	ctx    context.Context
	cancel context.CancelFunc
}

func NewChainWatcher(chainState *blockchain.ChainState, eventbus *eventbus.EventBus) *ChainWatcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &ChainWatcher{
		chainState: chainState,
		eventbus:   eventbus,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (cw *ChainWatcher) Start() {
	logger.Info("Starting ChainWatcher...")

	go cw.watchNewBlocks()
	go cw.watchFinalizedBlocks()
	go cw.watchSyncState()

	logger.Info("ChainWatcher started.")
}

func (cw *ChainWatcher) Stop() {
	logger.Info("Stopping ChainWatcher...")
	cw.cancel()
}

func (cw *ChainWatcher) watchNewBlocks() {
	logger.Info("Best block watcher started.")
	defer logger.Info("Best block watcher stopped.")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := cw.checkBestBlock(); err != nil {
				logger.Error("Error checking best block:", err)
			}
		case <-cw.ctx.Done():
			return
		}
	}
}

func (cw *ChainWatcher) checkBestBlock() error {
	latestBlock := cw.chainState.GetLatestBlock()

	currentHash, err := hash.ComputeBlockHeaderHash(latestBlock.Header)
	if err != nil {
		return fmt.Errorf("failed to compute block header hash: %v", err)
	}

	if currentHash != cw.latestBestBlockHash {
		logger.Info(fmt.Sprintf("New best block detected: 0x%x", currentHash[:8]))
		cw.latestBestBlockHash = currentHash

		cw.eventbus.Publish(eventbus.Event{
			Type: eventbus.EventNewBlock,
			Data: eventbus.BlockEvent{
				HeaderHash: encodeHash(currentHash),
				Slot:       uint64(latestBlock.Header.Slot),
			},
		})
	}
	return nil
}

func (cw *ChainWatcher) watchFinalizedBlocks() {
	logger.Info("Finalized block watcher started.")
	defer logger.Info("Finalized block watcher stopped.")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := cw.checkFinalizedBlock(); err != nil {
				logger.Error("Error checking finalized block:", err)
			}
		case <-cw.ctx.Done():
			return
		}
	}
}

func (cw *ChainWatcher) checkFinalizedBlock() error {
	finalizedBlocks := cw.chainState.GetFinalizedBlocks()

	var currentHash types.HeaderHash
	var currentSlot types.TimeSlot

	if len(finalizedBlocks) == 0 {
		genesisBlock := cw.chainState.GetGenesisBlock()

		var err error
		currentHash, err = hash.ComputeBlockHeaderHash(genesisBlock.Header)
		if err != nil {
			return fmt.Errorf("failed to compute genesis block header hash: %v", err)
		}
		currentSlot = genesisBlock.Header.Slot
	} else {
		lastFinalized := finalizedBlocks[len(finalizedBlocks)-1]

		var err error
		currentHash, err = hash.ComputeBlockHeaderHash(lastFinalized.Header)
		if err != nil {
			return fmt.Errorf("failed to compute finalized block header hash: %v", err)
		}
		currentSlot = lastFinalized.Header.Slot
	}

	if currentHash != cw.latestFinalizedBlockHash {
		logger.Info(fmt.Sprintf("New finalized block detected: 0x%x", currentHash[:8]))
		cw.latestFinalizedBlockHash = currentHash

		cw.eventbus.Publish(eventbus.Event{
			Type: eventbus.EventFinalizedBlock,
			Data: eventbus.BlockEvent{
				HeaderHash: encodeHash(currentHash),
				Slot:       uint64(currentSlot),
			},
		})
	}

	return nil
}

func (cw *ChainWatcher) watchSyncState() {
	logger.Info("Sync state watcher started.")
	defer logger.Info("Sync state watcher stopped.")

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := cw.checkSyncState(); err != nil {
				logger.Error("Error checking sync state:", err)
			}
		case <-cw.ctx.Done():
			return
		}
	}
}

func (cw *ChainWatcher) checkSyncState() error {
	// TODO: Implement actual sync state retrieval logic
	currentState := "Completed"
	numPeers := 0 // TODO: Retrieve actual number of peers

	if currentState != cw.latestSyncState {
		logger.Info(fmt.Sprintf("Sync state changed: %s", currentState))
		cw.latestSyncState = currentState

		cw.eventbus.Publish(eventbus.Event{
			Type: eventbus.EventSyncStateChanged,
			Data: eventbus.SyncStateEvent{
				NumPeers: numPeers,
				Status:   currentState,
			},
		})
	}
	return nil
}
