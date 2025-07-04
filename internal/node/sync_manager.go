package node

import (
	"context"
	"log"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// HeadInfo is now defined in quic package, import it if needed
type HeadInfo = quic.HeadInfo

// public typealias Data32 = FixedSizeData<ConstInt32> // golang -> types.OpaqueHash or types.ByteArray32
// public typealias TimeslotIndex = UInt32	// golang -> types.TimeSlot
// NetAddr // golang -> quic.Connction.RemoteAddr()

type SyncStatus int

const (
	Discovering SyncStatus = iota
	BulkSyncing
	Syncing
)

type SyncManager struct {
	ctx                  context.Context
	cancel               context.CancelFunc
	peers                map[string]*quic.Peer // Store quic.Peer directly
	eventBus             *quic.EventBus
	subscriptions        []quic.EventType
	blockchain           blockchain.Blockchain
	status               SyncStatus
	networkBest          *HeadInfo // Optional network best block
	networkFinalizedBest *HeadInfo // Optional network finalized best block
}

const BLOCK_REQUEST_BLOCK_COUNT uint32 = 50

func NewSyncManager() *SyncManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &SyncManager{
		ctx:      ctx,
		cancel:   cancel,
		peers:    make(map[string]*quic.Peer),
		eventBus: quic.NewEventBus(),
		status:   Discovering,
		// TODO: add blockchain
	}
}

func (sm *SyncManager) setupEventSubscriptions() {
	sm.eventBus.Subscribe(quic.PeerAdded, sm.handlePeerAdded)

	// newBlockHeader should be in handlePeerUpdated event
	sm.eventBus.Subscribe(quic.PeerUpdated, sm.handlePeerUpdated)
}

func (sm *SyncManager) handlePeerAdded(ctx context.Context, event quic.Event) error {
	if peerEvent, ok := event.(*quic.PeerAddedEvent); ok {
		return sm.onPeerUpdated(peerEvent.Peer, nil)
	}
	return nil
}

func (sm *SyncManager) handlePeerUpdated(ctx context.Context, event quic.Event) error {
	if peerEvent, ok := event.(*quic.PeerUpdatedEvent); ok {
		return sm.onPeerUpdated(peerEvent.Peer, peerEvent.NewBlockHeader)
	}
	return nil
}

// onPeerUpdated handles peer updates similar to the Swift implementation
func (sm *SyncManager) onPeerUpdated(peer *quic.Peer, newBlockHeader *HeadInfo) error {
	// TODO: improve this to handle the case misbehaved peers sending us the wrong best
	log.Printf("on peer updated: peer=%s, best=%+v, finalized=%+v, newBlockHeader=%+v",
		peer.ID, peer.Best, peer.Finalized, newBlockHeader)

	// Update network best
	if sm.networkBest != nil {
		if peer.Best != nil && peer.Best.Timeslot > sm.networkBest.Timeslot {
			sm.networkBest = peer.Best
		}
	} else {
		sm.networkBest = peer.Best
	}

	// Update network finalized best
	if sm.networkFinalizedBest != nil {
		if peer.Finalized.Timeslot > sm.networkFinalizedBest.Timeslot {
			sm.networkFinalizedBest = &peer.Finalized
		}
	} else {
		sm.networkFinalizedBest = &peer.Finalized
	}

	// Store peer
	sm.peers[peer.ID] = peer

	// Get current head from blockchain
	currentHead, err := sm.getCurrentHead()
	if err != nil {
		log.Printf("Error getting current head: %v", err)
		return err
	}

	// Check if sync is completed
	if sm.networkBest != nil && currentHead.Timeslot >= sm.networkBest.Timeslot {
		sm.syncCompleted()
		return nil
	}

	// Handle different sync states
	switch sm.status {
	case Discovering:
		sm.status = BulkSyncing
		return sm.bulkSync(currentHead)
	case BulkSyncing:
		return sm.bulkSync(currentHead)
	case Syncing:
		if newBlockHeader != nil {
			return sm.importBlock(currentHead.Timeslot, newBlockHeader, peer.ID)
		}
	}

	return nil
}

// getCurrentHead gets the current best head from the blockchain
func (sm *SyncManager) getCurrentHead() (*HeadInfo, error) {
	// TODO: implement real blockchain query
	// For now, return genesis block info
	head, err := sm.blockchain.GetCurrentHead()
	if err != nil {
		return nil, err
	}
	return &HeadInfo{
		Hash:     head.Header.Parent,
		Timeslot: head.Header.Slot,
	}, nil
}

// syncCompleted handles the completion of synchronization
func (sm *SyncManager) syncCompleted() {
	sm.status = Syncing
	// TODO: notify waiting goroutines (equivalent to Swift's continuation)
}

// bulkSync performs bulk synchronization
func (sm *SyncManager) bulkSync(currentHead *HeadInfo) error {
	log.Printf("Starting bulk sync from timeslot %d", currentHead.Timeslot)

	if sm.networkBest == nil {
		return nil
	}

	// TODO: implement actual bulk sync with block requests
	return nil
}

// importBlock imports a single block during normal sync
func (sm *SyncManager) importBlock(currentTimeslot types.TimeSlot, newHeader *HeadInfo, peerID string) error {
	log.Printf("Importing block from peer %s: timeslot=%d, hash=%x",
		peerID, newHeader.Timeslot, newHeader.Hash)
	// TODO: implement actual block import logic
	return nil
}

func (sm *SyncManager) Close() {
	sm.cancel()

	for _, sub := range sm.subscriptions {
		sm.eventBus.Unsubscribe(sub)
	}

	sm.subscriptions = nil
}
