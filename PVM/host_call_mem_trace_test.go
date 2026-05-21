//go:build trace

package PVM

import (
	"encoding/json"
	"testing"
)

type stubGuestMemory struct {
	readable func(addr, length uint64) bool
}

func (s *stubGuestMemory) IsReadable(addr, length uint64) bool {
	if s.readable != nil {
		return s.readable(addr, length)
	}
	return true
}
func (s *stubGuestMemory) IsWriteable(addr, length uint64) bool { return true }
func (s *stubGuestMemory) Read(addr, length uint64) []byte       { return make([]byte, length) }
func (s *stubGuestMemory) Write(addr uint64, data []byte)        {}

func TestBeginHostCallMemTraceRecordsAddrAndLenOnly(t *testing.T) {
	detailsFn, wrapped := BeginHostCallMemTrace(&stubGuestMemory{})

	wrapped.Read(0x32960, 8)
	wrapped.Write(0x32e40, []byte{1, 2, 3})

	raw := detailsFn()
	if raw == nil {
		t.Fatal("expected details JSON")
	}

	var got hostCallMemDetails
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got.MemReads) != 1 || got.MemReads[0].Addr != "0x32960" || got.MemReads[0].Len != 8 || !got.MemReads[0].Ok {
		t.Fatalf("memreads: %+v", got.MemReads)
	}
	if len(got.MemWrites) != 1 || got.MemWrites[0].Addr != "0x32e40" || got.MemWrites[0].Len != 3 || !got.MemWrites[0].Ok {
		t.Fatalf("memwrites: %+v", got.MemWrites)
	}
}

func TestBeginHostCallMemTraceRecordsFailedIsReadable(t *testing.T) {
	inner := &stubGuestMemory{
		readable: func(addr, length uint64) bool {
			return addr != 0x32e40
		},
	}
	detailsFn, wrapped := BeginHostCallMemTrace(inner)

	if wrapped.IsReadable(0x32960, 8) != true {
		t.Fatal("first check should pass")
	}
	if wrapped.IsReadable(0x32e40, 1104) != false {
		t.Fatal("second check should fail")
	}

	raw := detailsFn()
	var got hostCallMemDetails
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got.MemReads) != 2 {
		t.Fatalf("expected 2 read probes, got %+v", got.MemReads)
	}
	if !got.MemReads[0].Ok || got.MemReads[0].Addr != "0x32960" {
		t.Fatalf("first probe: %+v", got.MemReads[0])
	}
	if got.MemReads[1].Ok || got.MemReads[1].Addr != "0x32e40" || got.MemReads[1].Len != 1104 {
		t.Fatalf("second probe: %+v", got.MemReads[1])
	}
}
