//go:build trace

package PVMtrace

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Trace is the active trace recorder (trace build only).
type Trace struct {
	cfg     TraceConfig
	dir     string
	info    TraceInfo
	steps   int64
	closed  bool

	pcW      *streamWriter
	opcodeW  *streamWriter
	gasW     *streamWriter
	dstValW  *streamWriter
	src1ValW *streamWriter
	src2ValW *streamWriter
	loadsW   *streamWriter
	storesW  *streamWriter
	hostCallW *streamWriter

	// reusable buffers
	buf8  [8]byte
	buf12 [12]byte
}

// NewTrace creates a trace directory and opens all stream writers.
// Returns nil if JAM_PVM_TRACE_DIR is not set.
func NewTrace(serviceID uint32, codeHash [32]byte, timeslot uint64, programCode []byte, init InitialState, invocationType, backend string) *Trace {
	cfg := LoadConfigFromEnv()
	if cfg.Dir == "" {
		return nil
	}

	codeHashHex := fmt.Sprintf("%x", codeHash[:8])
	dirName := fmt.Sprintf("%d_%s", serviceID, codeHashHex)
	dir := cfg.Dir
	if cfg.RunID != "" {
		dir = filepath.Join(dir, cfg.RunID)
	}
	dir = filepath.Join(dir, dirName)

	if err := os.MkdirAll(filepath.Join(dir, "meta"), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "pvmtrace: mkdir %s: %v\n", dir, err)
		return nil
	}

	traceMode := TraceModeNormal
	if backend == BackendRecompiler {
		traceMode = TraceModeDebugSingleStep
	}

	t := &Trace{
		cfg: cfg,
		dir: dir,
		info: TraceInfo{
			FormatVersion:    FormatVersion,
			GraypaperVersion: GraypaperVersion,
			Backend:          backend,
			TraceMode:        traceMode,
			ServiceID:        serviceID,
			CodeHash:         fmt.Sprintf("0x%x", codeHash[:]),
			Timeslot:         timeslot,
			InvocationType:   invocationType,
			RunID:            cfg.RunID,
			InitialPC:        init.PC,
			InitialGas:       init.Gas,
			InitialRegs:      init.Regs,
		},
	}

	if len(programCode) > 0 {
		progPath := filepath.Join(dir, "meta", "program.bin")
		os.WriteFile(progPath, programCode, 0644)
	}

	var err error
	openW := func(name string) *streamWriter {
		if err != nil {
			return nil
		}
		var w *streamWriter
		w, err = newStreamWriter(dir, name, cfg.GzipLevel, cfg.BufferMB)
		return w
	}

	t.pcW = openW(StreamPC)
	t.opcodeW = openW(StreamOpcode)
	t.gasW = openW(StreamGas)
	t.dstValW = openW(StreamDstVal)
	t.src1ValW = openW(StreamSrc1Val)
	t.src2ValW = openW(StreamSrc2Val)
	t.loadsW = openW(StreamLoads)
	t.storesW = openW(StreamStores)
	t.hostCallW = openW(StreamHostCalls)

	if err != nil {
		fmt.Fprintf(os.Stderr, "pvmtrace: open streams: %v\n", err)
		t.closeWriters()
		return nil
	}

	t.info.Streams = AllStreamNames
	return t
}

// Steps returns the number of instruction steps recorded so far.
func (t *Trace) Steps() int64 {
	if t == nil {
		return 0
	}
	return t.steps
}

// RecordStep records one instruction step to all per-instruction streams.
func (t *Trace) RecordStep(pc uint32, opcode byte, dst, src0, src1 uint8, dstVal, src1Val, src2Val uint64, gas int64, loadAddr uint32, loadVal uint64, storeAddr uint32, storeVal uint64) {
	if t == nil || t.closed {
		return
	}
	if t.cfg.MaxSteps > 0 && t.steps >= t.cfg.MaxSteps {
		t.info.Truncated = true
		return
	}

	binary.LittleEndian.PutUint64(t.buf8[:], uint64(pc))
	t.pcW.Write(t.buf8[:])

	t.opcodeW.Write([]byte{opcode})

	binary.LittleEndian.PutUint64(t.buf8[:], uint64(gas))
	t.gasW.Write(t.buf8[:])

	binary.LittleEndian.PutUint64(t.buf8[:], dstVal)
	t.dstValW.Write(t.buf8[:])

	binary.LittleEndian.PutUint64(t.buf8[:], src1Val)
	t.src1ValW.Write(t.buf8[:])

	binary.LittleEndian.PutUint64(t.buf8[:], src2Val)
	t.src2ValW.Write(t.buf8[:])

	binary.LittleEndian.PutUint32(t.buf12[0:4], loadAddr)
	binary.LittleEndian.PutUint64(t.buf12[4:12], loadVal)
	t.loadsW.Write(t.buf12[:])

	binary.LittleEndian.PutUint32(t.buf12[0:4], storeAddr)
	binary.LittleEndian.PutUint64(t.buf12[4:12], storeVal)
	t.storesW.Write(t.buf12[:])

	t.steps++
}

// RecordBlockEntry is a no-op placeholder for production recompiler compatibility.
// Block-level metadata is not part of per-instruction streams.
func (t *Trace) RecordBlockEntry(pc uint64, gas int64) {}

// RecordHostCall writes one host-call boundary record to host_calls.jsonl.gz.
func (t *Trace) RecordHostCall(rec HostCallRecord) {
	if t == nil || t.closed || t.hostCallW == nil {
		return
	}
	data, err := json.Marshal(rec)
	if err != nil {
		return
	}
	data = append(data, '\n')
	t.hostCallW.Write(data)
}

// Close finalizes all streams and writes meta/info.json.
func (t *Trace) Close(final FinalState) error {
	if t == nil || t.closed {
		return nil
	}
	t.closed = true
	t.info.FinalPC = final.PC
	t.info.FinalGas = final.Gas
	t.info.FinalExitReason = final.ExitReason
	t.info.FinalRegs = final.Regs
	t.info.TotalSteps = t.steps

	t.closeWriters()

	infoData, err := json.MarshalIndent(t.info, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal info.json: %w", err)
	}
	infoPath := filepath.Join(t.dir, "meta", "info.json")
	return os.WriteFile(infoPath, infoData, 0644)
}

func (t *Trace) closeWriters() {
	for _, w := range []*streamWriter{
		t.pcW, t.opcodeW, t.gasW,
		t.dstValW, t.src1ValW, t.src2ValW,
		t.loadsW, t.storesW, t.hostCallW,
	} {
		if w != nil {
			w.Close()
		}
	}
}
