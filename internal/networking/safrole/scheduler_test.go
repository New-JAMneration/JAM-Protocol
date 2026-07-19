package safrole

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/networking/epochclock"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/stretchr/testify/require"
)

func TestScheduler_gatesStep1AndStep2(t *testing.T) {
	backup := types.EpochLength
	t.Cleanup(func() { types.EpochLength = backup })
	types.EpochLength = 60

	s := NewScheduler(nil)
	applied := epochclock.ConnectivityApplied{
		Epoch:          1,
		EpochStartSlot: 60,
		AppliedAtSlot:  60,
	}
	s.mu.Lock()
	s.connectivity = &applied
	s.mu.Unlock()

	require.False(t, s.CanSendStep1(60), "E/60 delay not met at slot 60")
	require.True(t, s.CanSendStep1(61))

	require.False(t, s.CanForwardStep2(62), "E/20 delay not met at slot 62")
	require.True(t, s.CanForwardStep2(63), "slot 63 is 3 slots after applied 60")
}

func TestScheduler_eventBusOpensStep1Window(t *testing.T) {
	backup := types.EpochLength
	t.Cleanup(func() { types.EpochLength = backup })
	types.EpochLength = 60

	var finalized types.TimeSlot = 59
	bus := quic.NewEventBus()
	s := NewScheduler(nil)
	s.finalizedOverride = func() types.TimeSlot { return finalized }
	s.Start(bus)

	require.NoError(t, bus.PublishConnectivityApplied(context.Background(), epochclock.ConnectivityApplied{
		Epoch:          1,
		EpochStartSlot: 60,
		AppliedAtSlot:  60,
	}))

	require.False(t, s.CanSendStep1(finalized))

	finalized = 65
	require.NoError(t, bus.PublishBlockImported(context.Background(), quic.HeadInfo{Timeslot: finalized}))

	require.True(t, s.CanSendStep1(finalized))
}

func TestScheduler_WaitForwardStep2UnblocksOnImport(t *testing.T) {
	backup := types.EpochLength
	t.Cleanup(func() { types.EpochLength = backup })
	types.EpochLength = 60

	var finalized atomic.Uint32
	finalized.Store(60)
	s := NewScheduler(nil)
	s.finalizedOverride = func() types.TimeSlot { return types.TimeSlot(finalized.Load()) }
	s.mu.Lock()
	s.connectivity = &epochclock.ConnectivityApplied{
		Epoch:          1,
		EpochStartSlot: 60,
		AppliedAtSlot:  60,
	}
	s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- s.WaitForwardStep2(ctx) }()

	time.Sleep(50 * time.Millisecond)
	select {
	case err := <-done:
		t.Fatalf("wait returned early: %v", err)
	default:
	}

	finalized.Store(63)
	s.notifySlotAdvance()

	require.NoError(t, <-done)
}

func TestScheduler_WaitForwardStep2BroadcastsToMultipleWaiters(t *testing.T) {
	backup := types.EpochLength
	t.Cleanup(func() { types.EpochLength = backup })
	types.EpochLength = 60

	var finalized atomic.Uint32
	finalized.Store(60)

	s := NewScheduler(nil)
	s.finalizedOverride = func() types.TimeSlot {
		return types.TimeSlot(finalized.Load())
	}
	s.mu.Lock()
	s.connectivity = &epochclock.ConnectivityApplied{
		Epoch:          1,
		EpochStartSlot: 60,
		AppliedAtSlot:  60,
	}
	s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	const waiters = 8
	done := make(chan error, waiters)
	for i := 0; i < waiters; i++ {
		go func() {
			done <- s.WaitForwardStep2(ctx)
		}()
	}

	time.Sleep(50 * time.Millisecond)
	for i := 0; i < waiters; i++ {
		select {
		case err := <-done:
			t.Fatalf("waiter %d returned early: %v", i, err)
		default:
		}
	}

	finalized.Store(63)
	s.notifySlotAdvance()

	for i := 0; i < waiters; i++ {
		require.NoError(t, <-done)
	}
}
