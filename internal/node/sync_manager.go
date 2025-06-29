package node

import (
	"context"

	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
)

// public struct HeadInfo: Sendable {
//     public var hash: Data32
//     public var timeslot: TimeslotIndex
//     public var number: UInt32
// }

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
	ctx           context.Context
	cancel        context.CancelFunc
	peers         map[string]struct{}
	eventBus      *quic.EventBus
	subscriptions []quic.EventType
}

// const BLOCK_REQUEST_BLOCK_COUNT uint32 = 50

func NewSyncManager() *SyncManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &SyncManager{
		ctx:      ctx,
		cancel:   cancel,
		peers:    make(map[string]struct{}),
		eventBus: quic.NewEventBus(),
	}
}

func (sm *SyncManager) setupEventSubscriptions() {
	sm.eventBus.Subscribe(quic.PeerAdded, sm.handlePeerAdded)
	sm.eventBus.Subscribe(quic.PeerUpdated, sm.handlePeerUpdated)
}

func (sm *SyncManager) handlePeerAdded(ctx context.Context, event quic.Event) error {
	if _, ok := event.(quic.Peer); ok {
		// return sm.onPeerUpdated(peer)
	}
	return nil
}

func (sm *SyncManager) handlePeerUpdated(ctx context.Context, event quic.Event) error {
	if _, ok := event.(quic.Peer); ok {
		// return sm.onPeerUpdated(peer)
	}
	return nil
}

func (sm *SyncManager) Close() {
	sm.cancel()

	for _, sub := range sm.subscriptions {
		sm.eventBus.Unsubscribe(sub)
	}

	sm.subscriptions = nil
}
