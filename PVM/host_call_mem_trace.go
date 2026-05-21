//go:build trace

package PVM

import (
	"encoding/json"
	"fmt"
)

// hostCallMemAccess records a guest-memory probe or access during a host-call.
// IsReadable/IsWriteable append a check entry; Read/Write append an access entry.
type hostCallMemAccess struct {
	Addr string `json:"addr"`
	Len  int    `json:"len"`
	Ok   bool   `json:"ok"`
}

// hostCallMemDetails is written into HostCallRecord.details (PVMtrace schema).
type hostCallMemDetails struct {
	MemReads  []hostCallMemAccess `json:"memreads,omitempty"`
	MemWrites []hostCallMemAccess `json:"memwrites,omitempty"`
}

type hostCallMemTracer struct {
	inner  GuestMemory
	reads  []hostCallMemAccess
	writes []hostCallMemAccess
}

// BeginHostCallMemTrace wraps inner for one omega invocation. Install wrapped on
// VMState.Mem, then call detailsFn after omega returns.
func BeginHostCallMemTrace(inner GuestMemory) (detailsFn func() json.RawMessage, wrapped GuestMemory) {
	t := &hostCallMemTracer{inner: inner}
	return t.detailsJSON, t
}

func (t *hostCallMemTracer) IsReadable(addr, length uint64) bool {
	ok := t.inner.IsReadable(addr, length)
	t.reads = append(t.reads, hostCallMemAccess{
		Addr: fmt.Sprintf("0x%x", addr),
		Len:  int(length),
		Ok:   ok,
	})
	return ok
}

func (t *hostCallMemTracer) IsWriteable(addr, length uint64) bool {
	ok := t.inner.IsWriteable(addr, length)
	t.writes = append(t.writes, hostCallMemAccess{
		Addr: fmt.Sprintf("0x%x", addr),
		Len:  int(length),
		Ok:   ok,
	})
	return ok
}

func (t *hostCallMemTracer) Read(addr, length uint64) []byte {
	t.reads = append(t.reads, hostCallMemAccess{
		Addr: fmt.Sprintf("0x%x", addr),
		Len:  int(length),
		Ok:   true,
	})
	return t.inner.Read(addr, length)
}

func (t *hostCallMemTracer) Write(addr uint64, data []byte) {
	t.writes = append(t.writes, hostCallMemAccess{
		Addr: fmt.Sprintf("0x%x", addr),
		Len:  len(data),
		Ok:   true,
	})
	t.inner.Write(addr, data)
}

func (t *hostCallMemTracer) detailsJSON() json.RawMessage {
	if len(t.reads) == 0 && len(t.writes) == 0 {
		return nil
	}
	d := hostCallMemDetails{
		MemReads:  t.reads,
		MemWrites: t.writes,
	}
	b, err := json.Marshal(d)
	if err != nil {
		return nil
	}
	return b
}
