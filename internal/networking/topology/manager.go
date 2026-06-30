package topology

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"log"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/validator"
	"github.com/New-JAMneration/JAM-Protocol/internal/safrole"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

const defaultReconcileEvery = 15 * time.Second

// Manager maintains required validator transport connectivity and applies epoch-transition delays.
type Manager struct {
	peer     *quic.Peer
	vm       *validator.ValidatorManager
	chain    *blockchain.ChainState
	selfKey  types.Ed25519Public
	eventBus *quic.EventBus

	pendingEpoch *pendingEpochTransition
	appliedEpoch types.TimeSlot
}

type pendingEpochTransition struct {
	targetEpoch    types.TimeSlot
	epochStartSlot types.TimeSlot
}

// NewManager creates a topology manager for a validator node.
func NewManager(peer *quic.Peer, vm *validator.ValidatorManager, chain *blockchain.ChainState, selfKey types.Ed25519Public) *Manager {
	return &Manager{
		peer:    peer,
		vm:      vm,
		chain:   chain,
		selfKey: selfKey,
	}
}

// Start wires sync-driven reconcile hooks and runs the periodic reconcile loop.
func (m *Manager) Start(ctx context.Context, eventBus *quic.EventBus) {
	if m == nil {
		return
	}
	m.eventBus = eventBus
	if eventBus != nil {
		eventBus.Subscribe(quic.BlockImported, m.handleBlockImported)
	}
	go m.Run(ctx)
}

func (m *Manager) handleBlockImported(ctx context.Context, event quic.Event) error {
	imported, ok := event.(*quic.BlockImportedEvent)
	if !ok || imported == nil {
		return nil
	}
	ts := readChainTimeslots(m.chain)
	log.Printf("topology: block imported at slot %d (chain head %d, finalized %d)",
		imported.Head.Timeslot, ts.BestHead, ts.Finalized)
	m.reconcile(ctx)
	return nil
}

// CurrentTimeslots returns best-head and finalized slots from local chain state.
func (m *Manager) CurrentTimeslots() ChainTimeslots {
	return readChainTimeslots(m.chain)
}

// Run periodically reconciles transport connectivity until ctx is cancelled.
func (m *Manager) Run(ctx context.Context) {
	if m == nil || m.peer == nil || m.vm == nil || m.chain == nil {
		return
	}
	ticker := time.NewTicker(defaultReconcileEvery)
	defer ticker.Stop()

	m.reconcile(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.reconcile(ctx)
		}
	}
}

// ReconcileOnce refreshes validator sets, applies any pending epoch transition, and dials/prunes peers.
func (m *Manager) ReconcileOnce(ctx context.Context) {
	m.reconcile(ctx)
}

func (m *Manager) reconcile(ctx context.Context) {
	finalized := finalizedBlock(m.chain)
	m.trackEpochTransition(finalized)

	if m.pendingEpoch != nil && !m.canApplyEpochTransition(finalized) {
		targets := validator.TransportTargets(m.vm.Grid, m.vm.SelfIndex)
		m.dialMissing(ctx, targets)
		return
	}
	if m.pendingEpoch != nil {
		log.Printf("topology: applying epoch %d connectivity after delay", m.pendingEpoch.targetEpoch)
		m.appliedEpoch = m.pendingEpoch.targetEpoch
		m.pendingEpoch = nil
	}

	m.vm.RefreshFromChain(m.chain)
	targets := validator.TransportTargets(m.vm.Grid, m.vm.SelfIndex)
	desired := desiredKeySet(targets, m.selfKey)

	m.dialMissing(ctx, targets)
	m.pruneStale(desired)
}

func (m *Manager) trackEpochTransition(finalized types.Block) {
	epoch := safrole.GetEpochIndex(finalized.Header.Slot)
	if m.appliedEpoch == 0 {
		m.appliedEpoch = epoch
		return
	}
	if epoch <= m.appliedEpoch {
		return
	}
	if m.pendingEpoch != nil && m.pendingEpoch.targetEpoch == epoch {
		return
	}
	m.pendingEpoch = &pendingEpochTransition{
		targetEpoch:    epoch,
		epochStartSlot: epoch * types.TimeSlot(types.EpochLength),
	}
}

func (m *Manager) canApplyEpochTransition(finalized types.Block) bool {
	if m.pendingEpoch == nil {
		return true
	}
	epoch := safrole.GetEpochIndex(finalized.Header.Slot)
	if epoch < m.pendingEpoch.targetEpoch {
		return false
	}
	delay := EpochTransitionDelaySlots()
	slotsSince := finalized.Header.Slot - m.pendingEpoch.epochStartSlot
	return slotsSince >= types.TimeSlot(delay)
}

// EpochTransitionDelaySlots returns max(floor(E/30), 1) per JAMNP-S epoch transitions.
func EpochTransitionDelaySlots() int {
	delay := int(types.EpochLength) / 30
	if delay < 1 {
		delay = 1
	}
	return delay
}

// TransportPeerKeys returns desired transport peer Ed25519 keys (for tests and diagnostics).
func (m *Manager) TransportPeerKeys() []types.Ed25519Public {
	if m.vm == nil {
		return nil
	}
	m.vm.RefreshFromChain(m.chain)
	targets := validator.TransportTargets(m.vm.Grid, m.vm.SelfIndex)
	keys := make([]types.Ed25519Public, 0, len(targets))
	for _, v := range targets {
		if v.Ed25519 != m.selfKey {
			keys = append(keys, v.Ed25519)
		}
	}
	return keys
}

// GossipPeerKeys returns grid-gossip-eligible peer keys (subset of transport connectivity).
func (m *Manager) GossipPeerKeys() []types.Ed25519Public {
	if m.vm == nil {
		return nil
	}
	return validator.GossipPeerKeys(m.vm)
}

func desiredKeySet(targets []types.Validator, selfKey types.Ed25519Public) map[string]struct{} {
	out := make(map[string]struct{}, len(targets))
	for _, v := range targets {
		if v.Ed25519 == selfKey {
			continue
		}
		out[keyHex(v.Ed25519)] = struct{}{}
	}
	return out
}

func (m *Manager) dialMissing(ctx context.Context, targets []types.Validator) {
	if ctx.Err() != nil {
		return
	}
	m.peer.StartValidatorConnections(ctx, targets, m.selfKey)
}

func (m *Manager) pruneStale(desired map[string]struct{}) {
	for _, key := range m.peer.ConnectedPeerKeys() {
		if len(key) != ed25519.PublicKeySize {
			continue
		}
		var typed types.Ed25519Public
		copy(typed[:], key)
		if keyEqual(typed, m.selfKey) {
			continue
		}
		if _, ok := desired[keyHex(typed)]; ok {
			continue
		}
		if err := m.peer.DisconnectPeer(key); err != nil {
			log.Printf("topology: disconnect stale peer %x: %v", key[:4], err)
		}
	}
}

func keyHex(key types.Ed25519Public) string {
	return hex.EncodeToString(key[:])
}

func keyEqual(a, b types.Ed25519Public) bool {
	return ed25519.PublicKey(a[:]).Equal(ed25519.PublicKey(b[:]))
}
