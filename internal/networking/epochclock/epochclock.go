package epochclock

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/safrole"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// Timeslots exposes best-head and finalized slots from local chain state.
type Timeslots struct {
	BestHead  types.TimeSlot
	Finalized types.TimeSlot
}

// TimeslotsFrom reads chain head and finalized slots.
func TimeslotsFrom(chain *blockchain.ChainState) Timeslots {
	if chain == nil {
		return Timeslots{}
	}
	ts := Timeslots{
		Finalized: chain.GetLatestFinalizedBlock().Header.Slot,
	}
	if head, err := chain.GetCurrentHead(); err == nil {
		ts.BestHead = head.Header.Slot
	}
	return ts
}

// ConnectivityApplied marks when JAMNP-S required connectivity was applied for an epoch.
// Safrole CE 131/132 and UP 0 stream adjustments anchor to AppliedAtSlot.
type ConnectivityApplied struct {
	Epoch          types.TimeSlot
	EpochStartSlot types.TimeSlot
	AppliedAtSlot  types.TimeSlot
}

// EpochTransitionDelaySlots returns max(floor(E/30), 1) per JAMNP-S epoch transitions.
func EpochTransitionDelaySlots() int {
	return maxSlotDelay(int(types.EpochLength) / 30)
}

// SafroleStep1DelaySlots returns max(floor(E/60), 1) for CE 131 first-step timing.
func SafroleStep1DelaySlots() int {
	return maxSlotDelay(int(types.EpochLength) / 60)
}

// SafroleStep2DelaySlots returns max(floor(E/20), 1) for CE 132 forwarding delay.
func SafroleStep2DelaySlots() int {
	return maxSlotDelay(int(types.EpochLength) / 20)
}

// CanApplyEpochTransition reports whether finalized slot meets JAMNP-S epoch-transition rules.
func CanApplyEpochTransition(finalized types.TimeSlot, targetEpoch, epochStartSlot types.TimeSlot) bool {
	epoch := safrole.GetEpochIndex(finalized)
	if epoch < targetEpoch {
		return false
	}
	slotsSince := finalized - epochStartSlot
	return slotsSince >= types.TimeSlot(EpochTransitionDelaySlots())
}

// CanSendSafroleStep1 reports whether CE 131 step-1 may begin after connectivity was applied.
func (a ConnectivityApplied) CanSendSafroleStep1(finalized types.TimeSlot) bool {
	return a.slotsSinceApplied(finalized) >= types.TimeSlot(SafroleStep1DelaySlots())
}

// CanForwardSafroleStep2 reports whether CE 132 forwarding may begin after connectivity was applied.
func (a ConnectivityApplied) CanForwardSafroleStep2(finalized types.TimeSlot) bool {
	return a.slotsSinceApplied(finalized) >= types.TimeSlot(SafroleStep2DelaySlots())
}

func (a ConnectivityApplied) slotsSinceApplied(finalized types.TimeSlot) types.TimeSlot {
	if finalized < a.AppliedAtSlot {
		return 0
	}
	return finalized - a.AppliedAtSlot
}

func maxSlotDelay(delay int) int {
	if delay < 1 {
		return 1
	}
	return delay
}
