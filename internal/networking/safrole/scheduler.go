package safrole

import (
	"context"
	"log"
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/epochclock"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// TimingGate controls when Safrole CE 132 forwarding may begin after connectivity is applied.
type TimingGate interface {
	WaitForwardStep2(ctx context.Context) error
}

// Scheduler gates Safrole CE 131/132 timing on ConnectivityApplied anchors from topology.
type Scheduler struct {
	chain    *blockchain.ChainState
	eventBus *quic.EventBus

	mu           sync.Mutex
	connectivity *epochclock.ConnectivityApplied
	step1Opened  bool
	slotAdvance  chan struct{}

	// finalizedOverride is set in tests to avoid mutating global chain state.
	finalizedOverride func() types.TimeSlot
}

// NewScheduler creates a Safrole timing scheduler for a validator node.
func NewScheduler(chain *blockchain.ChainState) *Scheduler {
	return &Scheduler{
		chain:       chain,
		slotAdvance: make(chan struct{}),
	}
}

// Start subscribes to connectivity and block-import events.
func (s *Scheduler) Start(eventBus *quic.EventBus) {
	if s == nil || eventBus == nil {
		return
	}
	s.eventBus = eventBus
	eventBus.Subscribe(quic.ConnectivityApplied, s.handleConnectivityApplied)
	eventBus.Subscribe(quic.BlockImported, s.handleBlockImported)
}

func (s *Scheduler) handleConnectivityApplied(ctx context.Context, event quic.Event) error {
	ev, ok := event.(*quic.ConnectivityAppliedEvent)
	if !ok || ev == nil {
		return nil
	}
	s.mu.Lock()
	s.connectivity = &ev.Applied
	s.step1Opened = false
	s.mu.Unlock()
	log.Printf("safrole: connectivity applied for epoch %d at slot %d",
		ev.Applied.Epoch, ev.Applied.AppliedAtSlot)
	s.notifySlotAdvance()
	return nil
}

func (s *Scheduler) handleBlockImported(ctx context.Context, event quic.Event) error {
	_ = ctx
	_, ok := event.(*quic.BlockImportedEvent)
	if !ok {
		return nil
	}
	s.maybeOpenStep1()
	s.notifySlotAdvance()
	return nil
}

func (s *Scheduler) maybeOpenStep1() {
	finalized := s.finalizedSlot()
	if !s.CanSendStep1(finalized) {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.step1Opened || s.connectivity == nil {
		return
	}
	s.step1Opened = true
	log.Printf("safrole: CE 131 step-1 window open (epoch %d, finalized slot %d)",
		s.connectivity.Epoch, finalized)
}

// ConnectivityAnchor returns the latest connectivity timing anchor.
func (s *Scheduler) ConnectivityAnchor() *epochclock.ConnectivityApplied {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.connectivity == nil {
		return nil
	}
	cp := *s.connectivity
	return &cp
}

// CanSendStep1 reports whether CE 131 step-1 may begin for the given finalized slot.
func (s *Scheduler) CanSendStep1(finalized types.TimeSlot) bool {
	anchor := s.ConnectivityAnchor()
	if anchor == nil {
		return false
	}
	return anchor.CanSendSafroleStep1(finalized)
}

// CanForwardStep2 reports whether CE 132 forwarding may begin for the given finalized slot.
func (s *Scheduler) CanForwardStep2(finalized types.TimeSlot) bool {
	anchor := s.ConnectivityAnchor()
	if anchor == nil {
		return false
	}
	return anchor.CanForwardSafroleStep2(finalized)
}

// WaitForwardStep2 blocks until CE 132 forwarding is allowed or ctx is cancelled.
func (s *Scheduler) WaitForwardStep2(ctx context.Context) error {
	if s == nil {
		return nil
	}
	for {
		finalized := s.finalizedSlot()
		s.mu.Lock()
		if s.connectivity != nil && s.connectivity.CanForwardSafroleStep2(finalized) {
			s.mu.Unlock()
			return nil
		}
		slotAdvance := s.slotAdvance
		s.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-slotAdvance:
		}
	}
}

func (s *Scheduler) finalizedSlot() types.TimeSlot {
	if s == nil {
		return 0
	}
	if s.finalizedOverride != nil {
		return s.finalizedOverride()
	}
	if s.chain == nil {
		return 0
	}
	return s.chain.GetLatestFinalizedBlock().Header.Slot
}

func (s *Scheduler) notifySlotAdvance() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	close(s.slotAdvance)
	s.slotAdvance = make(chan struct{})
}
