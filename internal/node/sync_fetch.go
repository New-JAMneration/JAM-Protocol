package node

import (
	"fmt"
	"log"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

const maxFetchPeerAttempts = 3

// fetchWithPeerFallback tries CE128 fetch from primary and up to maxFetchPeerAttempts peers.
func (sm *SyncManager) fetchWithPeerFallback(primary *quic.Peer, from types.HeaderHash, direction byte, maxBlocks uint32) (int, error) {
	candidates := sm.peerCandidates(primary)
	if len(candidates) == 0 {
		return 0, fmt.Errorf("no sync peers available")
	}
	if len(candidates) > maxFetchPeerAttempts {
		candidates = candidates[:maxFetchPeerAttempts]
	}

	var lastErr error
	for _, peer := range candidates {
		imported, err := sm.fetchAndStoreBlocks(peer, from, direction, maxBlocks)
		if err == nil {
			return imported, nil
		}
		log.Printf("sync: CE128 fetch from peer %s failed: %v", peerID(peer), err)
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("CE128 fetch failed for all peers")
	}
	return 0, lastErr
}

func (sm *SyncManager) peerCandidates(primary *quic.Peer) []*quic.Peer {
	seen := make(map[string]struct{})
	out := make([]*quic.Peer, 0, len(sm.peers)+1)

	add := func(peer *quic.Peer) {
		if peer == nil || len(peer.Ed25519Key) == 0 {
			return
		}
		id := peerID(peer)
		if id == "" {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		out = append(out, peer)
	}

	add(primary)
	for _, peer := range sm.peers {
		add(peer)
	}
	return out
}

// advanceHeadForTest simulates a successful import for integration tests.
func (sm *SyncManager) advanceHeadForTest(slot types.TimeSlot, head types.HeaderHash) {
	if sm == nil || sm.blockchain == nil {
		return
	}
	chain, ok := sm.blockchain.(*blockchain.ChainState)
	if !ok {
		return
	}
	block := types.Block{
		Header: types.Header{
			Slot:   slot,
			Parent: head,
		},
	}
	chain.AddBlock(block)
	blockHash, err := hash.ComputeBlockHeaderHash(block.Header)
	if err != nil {
		return
	}
	chain.SetCurrentHead(types.HeaderHash(blockHash))
}
