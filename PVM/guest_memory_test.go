package PVM

import "testing"

func TestHeapMaxPagesReservesMajorZone(t *testing.T) {
	stackStart := uint64(1<<32 - 2*ZZ - ZI)
	mem := &Memory{heapLimit: stackStart}
	got := NewPagedGuestMemory(mem).HeapMaxPages()
	want := (stackStart - uint64(ZZ)) / uint64(ZP)
	if got != want {
		t.Fatalf("HeapMaxPages = %d, want %d", got, want)
	}
}
