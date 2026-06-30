package node

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	cehandler "github.com/New-JAMneration/JAM-Protocol/internal/networking/handler/ce"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// HeadInfo is now defined in quic package, import it if needed
type HeadInfo = quic.HeadInfo

type SyncStatus int

const (
	Discovering SyncStatus = iota
	BulkSyncing
	Syncing
)

type SyncManager struct {
	ctx                  context.Context
	cancel               context.CancelFunc
	localPeer            *quic.Peer
	peers                map[string]*quic.Peer
	eventBus             *quic.EventBus
	subscriptions        []quic.EventType
	blockchain           blockchain.Blockchain
	status               SyncStatus
	networkBest          *HeadInfo
	networkFinalizedBest *HeadInfo
}

const blockRequestBlockCount uint32 = 50

func NewSyncManager(bc blockchain.Blockchain, eventBus *quic.EventBus, localPeer *quic.Peer) *SyncManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &SyncManager{
		ctx:        ctx,
		cancel:     cancel,
		localPeer:  localPeer,
		peers:      make(map[string]*quic.Peer),
		eventBus:   eventBus,
		blockchain: bc,
		status:     Discovering,
	}
}

func (sm *SyncManager) Start() {
	sm.setupEventSubscriptions()
}

func (sm *SyncManager) setupEventSubscriptions() {
	if sm.eventBus == nil {
		return
	}
	sm.eventBus.Subscribe(quic.PeerAdded, sm.handlePeerAdded)
	sm.subscriptions = append(sm.subscriptions, quic.PeerAdded)
	sm.eventBus.Subscribe(quic.PeerUpdated, sm.handlePeerUpdated)
	sm.subscriptions = append(sm.subscriptions, quic.PeerUpdated)
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

func (sm *SyncManager) onPeerUpdated(peer *quic.Peer, newBlockHeader *HeadInfo) error {
	if peer == nil {
		return nil
	}
	if newBlockHeader != nil {
		if peer.Best == nil || newBlockHeader.Timeslot > peer.Best.Timeslot {
			peer.Best = newBlockHeader
		}
		if sm.networkBest == nil || newBlockHeader.Timeslot > sm.networkBest.Timeslot {
			sm.networkBest = newBlockHeader
		}
	}

	log.Printf("on peer updated: peer=%s, best=%+v, finalized=%+v, newBlockHeader=%+v",
		peerID(peer), peer.Best, peer.Finalized, newBlockHeader)

	if peer.Best != nil {
		if sm.networkBest == nil || peer.Best.Timeslot > sm.networkBest.Timeslot {
			sm.networkBest = peer.Best
		}
	}

	if sm.networkFinalizedBest != nil {
		if peer.Finalized.Timeslot > sm.networkFinalizedBest.Timeslot {
			sm.networkFinalizedBest = &peer.Finalized
		}
	} else if peer.Finalized.Timeslot > 0 || peer.Finalized.Hash != (types.HeaderHash{}) {
		sm.networkFinalizedBest = &peer.Finalized
	}

	sm.peers[peerID(peer)] = peer

	currentHead, err := sm.getCurrentHead()
	if err != nil {
		log.Printf("Error getting current head: %v", err)
		return err
	}

	if newBlockHeader != nil && newBlockHeader.Timeslot > currentHead.Timeslot && sm.localPeer != nil {
		if err := sm.importBlock(peer, currentHead, newBlockHeader); err != nil {
			log.Printf("CE128 import from announcement: %v", err)
		} else {
			currentHead, err = sm.getCurrentHead()
			if err != nil {
				return err
			}
		}
	}

	if sm.networkBest != nil && currentHead.Timeslot >= sm.networkBest.Timeslot {
		sm.syncCompleted()
		return nil
	}

	if sm.status == Discovering {
		sm.status = BulkSyncing
	}
	if sm.status == BulkSyncing {
		return sm.bulkSync(peer, currentHead)
	}

	return nil
}

func (sm *SyncManager) getCurrentHead() (*HeadInfo, error) {
	if sm.blockchain == nil {
		return nil, fmt.Errorf("sync manager blockchain dependency is nil")
	}
	head, err := sm.blockchain.GetCurrentHead()
	if err != nil {
		return nil, err
	}
	headerHash, err := hash.ComputeBlockHeaderHash(head.Header)
	if err != nil {
		return nil, err
	}
	return &HeadInfo{
		Hash:     headerHash,
		Timeslot: head.Header.Slot,
	}, nil
}

func (sm *SyncManager) syncCompleted() {
	sm.status = Syncing
}

func (sm *SyncManager) bulkSync(peer *quic.Peer, currentHead *HeadInfo) error {
	log.Printf("Starting bulk sync from timeslot %d", currentHead.Timeslot)
	if sm.networkBest == nil || sm.localPeer == nil {
		return nil
	}
	target := peer
	if target == nil || len(target.Ed25519Key) == 0 {
		target = sm.anyPeer()
	}
	if target == nil {
		return nil
	}
	return sm.fetchAndStoreBlocks(target, currentHead.Hash, 0, blockRequestBlockCount)
}

func (sm *SyncManager) importBlock(peer *quic.Peer, currentHead *HeadInfo, newHeader *HeadInfo) error {
	log.Printf("Importing block from peer %s: timeslot=%d, hash=%x",
		peerID(peer), newHeader.Timeslot, newHeader.Hash)
	if sm.localPeer == nil {
		return fmt.Errorf("local peer is nil")
	}
	if newHeader.Timeslot <= currentHead.Timeslot {
		return nil
	}
	maxBlocks := uint32(newHeader.Timeslot - currentHead.Timeslot)
	if maxBlocks > blockRequestBlockCount {
		maxBlocks = blockRequestBlockCount
	}
	if maxBlocks == 0 {
		maxBlocks = 1
	}
	return sm.fetchAndStoreBlocks(peer, currentHead.Hash, 0, maxBlocks)
}

func (sm *SyncManager) fetchAndStoreBlocks(peer *quic.Peer, from types.HeaderHash, direction byte, maxBlocks uint32) error {
	conn, ok := sm.localPeer.ConnectionFor(peer.Ed25519Key)
	if !ok {
		return fmt.Errorf("no connection for peer %s", peerID(peer))
	}

	blocks, err := cehandler.RequestBlocks(sm.ctx, conn, cehandler.CE128Payload{
		HeaderHash: from,
		Direction:  direction,
		MaxBlocks:  maxBlocks,
	})
	if err != nil {
		return err
	}
	return sm.storeBlocks(blocks)
}

func (sm *SyncManager) storeBlocks(blocks []types.Block) error {
	chain, ok := sm.blockchain.(*blockchain.ChainState)
	if !ok {
		return fmt.Errorf("blockchain does not support AddBlock")
	}
	for _, block := range blocks {
		chain.AddBlock(block)
	}
	return nil
}

func (sm *SyncManager) anyPeer() *quic.Peer {
	for _, peer := range sm.peers {
		return peer
	}
	return nil
}

func peerID(peer *quic.Peer) string {
	if peer == nil {
		return ""
	}
	if peer.ID != "" {
		return peer.ID
	}
	if len(peer.Ed25519Key) == ed25519.PublicKeySize {
		return hex.EncodeToString(peer.Ed25519Key)
	}
	return ""
}

func (sm *SyncManager) Close() {
	sm.cancel()

	if sm.eventBus == nil {
		sm.subscriptions = nil
		return
	}
	for _, sub := range sm.subscriptions {
		sm.eventBus.Unsubscribe(sub)
	}

	sm.subscriptions = nil
}
