//go:build linux && amd64 && cgo

package PVM_test

import (
	"bytes"
	"encoding/binary"
	"fmt"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/internal/service_account"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// Synthetic Ψ_M accumulate corpus (opt measurement; not jam-conformance).
//
//	Invoke
//	  entry        PC 0
//	  omegas       AccumulateOmegas
//	  gas budget   50_000_000
//	  service      id=1; StorageDict + PreimageLookup pre-filled; no unmatched keys
//
//	Standard-code layout (tiny mode)
//	  o            empty
//	  w            128 B  (keys / hash / dummy / log)
//	  z            0 extra RW pages
//	  s            4096 B stack
//	  rwStart      2·ZZ
//	  heapStart    rwStart + P(|w|) + z·ZP
//	  grow         n = h + extraPages, require h < n ≤ b
//	               h = heapStart/ZP
//	               b = (stackStart − ZZ)/ZP
//
//	Working set                         SmallData          LargeData
//	  grow extra pages                  2                  64
//	  mem-loop span                     256 B              extra·ZP (whole grown range)
//	  preimage / storage blob           16 B               4096 B
//	  fetch/lookup/read copy            64 B               4096 B
//
//	RW guest offsets (from rwStart)
//	  +0           storage key "k"
//	  +32          preimage hash (32 B)
//	  +64          dummy 32 B (yield / new / upgrade / …)
//	  +96          bless ServiceIDList (8 B)
//	  +104         log message (1 B)
//
//	Instruction skeleton (same for both sizes)
//	  1. grow_heap to n
//	  2. store/load u64 loop over [heapStart, heapStart+walk)
//	  3. host-call tour once each (see buildSynthInstructions)
//	  4. Halt with r7=r8=0
//
//	Host-call tour
//	  happy path   gas, grow_heap, fetch, lookup, read, write, info
//	  coverage     checkpoint, yield, query, log×1, bless, assign,
//	               designate, new, upgrade, transfer, eject,
//	               solicit, forget, provide
//	  not called   refine-only ids 7–14

const (
	synthEntryPC PVM.ProgramCounter = 0
	synthGas     types.Gas          = 50_000_000
	synthStack   uint32             = 4096
	synthZW      uint16             = 0
	synthService                    = types.ServiceID(1)

	synthOffKey   = 0
	synthOffHash  = 32
	synthOffDummy = 64
	synthOffBless = 96
	synthOffLog   = 104
	synthWLen     = 128

	synthSmallExtraPages = 2
	synthLargeExtraPages = 64
	synthSmallWalkBytes  = 256
	synthSmallBlobLen    = 16
	synthLargeBlobLen    = 4096
	synthSmallCopyLen    = 64
	synthLargeCopyLen    = 4096
)

type synthKind string

const (
	synthSmall synthKind = "SmallData"
	synthLarge synthKind = "LargeData"
)

type synthLayout struct {
	rwStart     uint32
	heapStart   uint32
	heapPages   uint64
	maxPages    uint64
	targetPages uint64
	walkBytes   uint64
	blobLen     int
	copyLen     uint64
}

type synthProgram struct {
	kind synthKind
	code []byte
	arg  PVM.Argument
	hash types.OpaqueHash
	lay  synthLayout
	pre  []byte
	stor []byte
}

func synthExtraPages(kind synthKind) uint64 {
	if kind == synthLarge {
		return synthLargeExtraPages
	}
	return synthSmallExtraPages
}

func computeSynthLayout(kind synthKind, wLen int) synthLayout {
	oLen := 0
	readWriteStart := 2*PVM.ZZ + PVM.Z(oLen)
	heapStart := readWriteStart + PVM.P(wLen) + uint32(synthZW)*PVM.ZP
	stackEnd := uint32(1<<32 - 2*PVM.ZZ - PVM.ZI)
	stackStart := stackEnd - PVM.P(int(synthStack))
	heapPages := uint64(heapStart) / uint64(PVM.ZP)
	maxPages := (uint64(stackStart) - PVM.ZZ) / uint64(PVM.ZP)
	extra := synthExtraPages(kind)
	target := heapPages + extra
	walk := uint64(synthSmallWalkBytes)
	blob := synthSmallBlobLen
	copyN := uint64(synthSmallCopyLen)
	if kind == synthLarge {
		walk = extra * uint64(PVM.ZP)
		blob = synthLargeBlobLen
		copyN = uint64(synthLargeCopyLen)
	}
	return synthLayout{
		rwStart:     readWriteStart,
		heapStart:   heapStart,
		heapPages:   heapPages,
		maxPages:    maxPages,
		targetPages: target,
		walkBytes:   walk,
		blobLen:     blob,
		copyLen:     copyN,
	}
}

func (lay synthLayout) growInRange() bool {
	return lay.targetPages > lay.heapPages && lay.targetPages <= lay.maxPages
}

func formatSynthInfo(p synthProgram) string {
	lay := p.lay
	return fmt.Sprintf(""+
		"synthetic data\n"+
		"  kind         %s\n"+
		"  code         %d B\n"+
		"  code_hash    0x%x\n"+
		"  entry        PC %d\n"+
		"  gas_budget   %d\n"+
		"  grow         h=%d  n=%d  b=%d  extra=%d  (h < n ≤ b)\n"+
		"  mem_walk     %d B\n"+
		"  blob         preimage=%d B  storage=%d B\n"+
		"  host_copy    %d B\n"+
		"  rw           start=0x%x  heap=0x%x  w=%d B",
		p.kind,
		len(p.code),
		p.hash[:],
		synthEntryPC,
		synthGas,
		lay.heapPages,
		lay.targetPages,
		lay.maxPages,
		lay.targetPages-lay.heapPages,
		lay.walkBytes,
		len(p.pre),
		len(p.stor),
		lay.copyLen,
		lay.rwStart,
		lay.heapStart,
		synthWLen,
	)
}

func wrapStandardCode(o, w, inner []byte, z uint16, s uint32) []byte {
	out := make([]byte, 0, 3+3+2+3+len(o)+len(w)+4+len(inner))
	out = append(out, encodeFixedLE(uint64(len(o)), 3)...)
	out = append(out, encodeFixedLE(uint64(len(w)), 3)...)
	out = append(out, encodeFixedLE(uint64(z), 2)...)
	out = append(out, encodeFixedLE(uint64(s), 3)...)
	out = append(out, o...)
	out = append(out, w...)
	out = append(out, encodeFixedLE(uint64(len(inner)), 4)...)
	out = append(out, inner...)
	return out
}

func encodeFixedLE(n uint64, nbytes int) []byte {
	b := make([]byte, nbytes)
	for i := range nbytes {
		b[i] = byte(n >> (8 * i))
	}
	return b
}

func buildPVMBlob(inst []byte, starts []int) []byte {
	bitmaskByteCount := len(inst) / 8
	if len(inst)%8 > 0 {
		bitmaskByteCount++
	}
	bitmask := make([]byte, bitmaskByteCount)
	for _, idx := range starts {
		bitmask[idx/8] |= 1 << (idx % 8)
	}
	var blob []byte
	blob = append(blob, 0) // jump table size (E_V < 0x80)
	blob = append(blob, 0) // jump table length
	if len(inst) < 0x80 {
		blob = append(blob, byte(len(inst)))
	} else {
		buf := make([]byte, 9)
		buf[0] = 0xFF
		binary.LittleEndian.PutUint64(buf[1:], uint64(len(inst)))
		blob = append(blob, buf...)
	}
	blob = append(blob, inst...)
	blob = append(blob, bitmask...)
	return blob
}

type pvmAsm struct {
	buf    []byte
	starts []int
}

func (a *pvmAsm) pc() int { return len(a.buf) }

func (a *pvmAsm) mark() { a.starts = append(a.starts, len(a.buf)) }

func packRegs(rA, rB uint8) byte { return (rB << 4) | (rA & 0x0F) }

func imm4u(v uint64) []byte {
	x := uint32(v)
	return []byte{byte(x), byte(x >> 8), byte(x >> 16), byte(x >> 24)}
}

func imm4s(v int32) []byte {
	u := uint32(v)
	return []byte{byte(u), byte(u >> 8), byte(u >> 16), byte(u >> 24)}
}

func (a *pvmAsm) ecalli(id uint8) {
	a.mark()
	a.buf = append(a.buf, 10, id)
}

func (a *pvmAsm) loadImm(reg uint8, val uint64) {
	a.mark()
	a.buf = append(a.buf, 51, reg&0x0F)
	a.buf = append(a.buf, imm4u(val)...)
}

func (a *pvmAsm) addImm64(dst, src uint8, imm uint64) {
	a.mark()
	a.buf = append(a.buf, 149, packRegs(dst, src))
	a.buf = append(a.buf, imm4u(imm)...)
}

func (a *pvmAsm) storeIndU64(valReg, addrReg uint8) {
	a.mark()
	a.buf = append(a.buf, 123, packRegs(valReg, addrReg), 0) // offset 0, 1-byte imm
}

func (a *pvmAsm) loadIndU64(dstReg, addrReg uint8) {
	a.mark()
	a.buf = append(a.buf, 130, packRegs(dstReg, addrReg), 0)
}

func (a *pvmAsm) branchLtU(rA, rB uint8, offset int32) {
	a.mark()
	a.buf = append(a.buf, 172, packRegs(rA, rB))
	a.buf = append(a.buf, imm4s(offset)...)
}

func (a *pvmAsm) fallthroughOp() {
	a.mark()
	a.buf = append(a.buf, 1)
}

func (a *pvmAsm) jumpInd(reg uint8) {
	a.mark()
	a.buf = append(a.buf, 50, reg&0x0F)
}

func buildSynthInstructions(lay synthLayout) ([]byte, []int) {
	// Host-call order after the mem loop (id → name):
	//   0 gas              2 fetch             3 lookup            4 read
	//   5 write            6 info             18 checkpoint       26 yield
	//  23 query          100 log              15 bless            16 assign
	//  17 designate       19 new              20 upgrade          21 transfer
	//  22 eject           24 solicit          25 forget           27 provide
	var a pvmAsm
	rw := uint64(lay.rwStart)
	heap := uint64(lay.heapStart)
	key := rw + synthOffKey
	hsh := rw + synthOffHash
	dummy := rw + synthOffDummy
	bless := rw + synthOffBless
	logMsg := rw + synthOffLog
	self := uint64(0xffffffff)

	a.loadImm(7, lay.targetPages)
	a.ecalli(1)

	a.loadImm(2, heap)
	a.loadImm(3, heap+lay.walkBytes)
	a.loadImm(4, 0x11)
	a.fallthroughOp()

	loopPC := a.pc()
	a.storeIndU64(4, 2)
	a.loadIndU64(5, 2)
	a.addImm64(2, 2, 8)
	branchPC := a.pc()
	a.branchLtU(2, 3, int32(loopPC-branchPC))

	a.ecalli(0)

	a.loadImm(10, 0)
	a.loadImm(7, heap)
	a.loadImm(8, 0)
	a.loadImm(9, lay.copyLen)
	a.ecalli(2)

	a.loadImm(7, self)
	a.loadImm(8, hsh)
	a.loadImm(9, heap)
	a.loadImm(10, 0)
	a.loadImm(11, lay.copyLen)
	a.ecalli(3)

	a.loadImm(7, self)
	a.loadImm(8, key)
	a.loadImm(9, 1)
	a.loadImm(10, heap)
	a.loadImm(11, 0)
	a.loadImm(12, lay.copyLen)
	a.ecalli(4)

	a.loadImm(7, key)
	a.loadImm(8, 1)
	a.loadImm(9, heap)
	a.loadImm(10, uint64(min(lay.blobLen, 16)))
	a.ecalli(5)

	a.loadImm(7, self)
	a.loadImm(8, heap)
	a.loadImm(9, 0)
	a.loadImm(10, 64)
	a.ecalli(6)

	a.ecalli(18)

	a.loadImm(7, dummy)
	a.ecalli(26)

	a.loadImm(7, hsh)
	a.loadImm(8, uint64(lay.blobLen))
	a.ecalli(23)

	a.loadImm(7, 3)
	a.loadImm(8, 0)
	a.loadImm(9, 0)
	a.loadImm(10, logMsg)
	a.loadImm(11, 1)
	a.ecalli(100)

	a.loadImm(7, 0)
	a.loadImm(8, bless)
	a.loadImm(9, 0)
	a.loadImm(10, 0)
	a.loadImm(11, heap)
	a.loadImm(12, 0)
	a.ecalli(15)

	a.loadImm(7, 99)
	a.loadImm(8, heap)
	a.loadImm(9, 0)
	a.ecalli(16)

	a.loadImm(7, dummy)
	a.loadImm(8, 0)
	a.ecalli(17)

	a.loadImm(7, dummy)
	a.loadImm(8, 0)
	a.loadImm(9, 0)
	a.loadImm(10, 0)
	a.loadImm(11, 1)
	a.loadImm(12, 0)
	a.ecalli(19)

	a.loadImm(7, dummy)
	a.loadImm(8, 0)
	a.loadImm(9, 0)
	a.ecalli(20)

	a.loadImm(7, 99)
	a.loadImm(8, 0)
	a.loadImm(9, 0)
	a.loadImm(10, heap)
	a.ecalli(21)

	a.loadImm(7, 99)
	a.loadImm(8, dummy)
	a.ecalli(22)

	a.loadImm(7, hsh)
	a.loadImm(8, uint64(lay.blobLen))
	a.ecalli(24)

	a.loadImm(7, hsh)
	a.loadImm(8, uint64(lay.blobLen))
	a.ecalli(25)

	a.loadImm(7, self)
	a.loadImm(8, key)
	a.loadImm(9, 1)
	a.ecalli(27)

	a.loadImm(7, 0)
	a.loadImm(8, 0)
	a.loadImm(0, 0xffff0000)
	a.jumpInd(0)

	return a.buf, a.starts
}

func makeSynthRW(preimage []byte) []byte {
	w := make([]byte, synthWLen)
	w[synthOffKey] = 'k'
	h := hash.Blake2bHash(preimage)
	copy(w[synthOffHash:synthOffHash+32], h[:])
	w[synthOffLog] = 'x'
	return w
}

func buildSynthProgram(kind synthKind) synthProgram {
	lay := computeSynthLayout(kind, synthWLen)
	if !lay.growInRange() {
		panic("synthetic grow_heap target is outside [h+1, b]")
	}
	pre := bytes.Repeat([]byte{0xab}, lay.blobLen)
	stor := bytes.Repeat([]byte{0xcd}, lay.blobLen)
	w := makeSynthRW(pre)
	inst, starts := buildSynthInstructions(lay)
	inner := buildPVMBlob(inst, starts)
	code := wrapStandardCode(nil, w, inner, synthZW, synthStack)
	return synthProgram{
		kind: kind,
		code: code,
		arg:  PVM.Argument{},
		hash: hash.Blake2bHash(code),
		lay:  lay,
		pre:  pre,
		stor: stor,
	}
}

func (p synthProgram) hostArgs() PVM.HostCallArgs {
	sid := synthService
	key := "k"
	items, octets := service_account.CalcStorageItemfootprint(key, p.stor)
	h := hash.Blake2bHash(p.pre)
	acct := types.ServiceAccount{
		ServiceInfo: types.ServiceInfo{
			CodeHash: p.hash,
			Balance:  1 << 60,
			Items:    items,
			Bytes:    octets,
		},
		PreimageLookup: types.PreimagesMapEntry{h: append(types.ByteSequence(nil), p.pre...)},
		LookupDict:     types.LookupMetaMapEntry{},
		StorageDict:    types.Storage{key: append(types.ByteSequence(nil), p.stor...)},
	}
	state := types.ServiceAccountState{sid: acct}
	storage := types.StateKeyVals{}
	partial := types.PartialStateSet{
		ServiceAccounts: state,
		AlwaysAccum:     types.AlwaysAccumulateMap{},
		Assign:          make(types.ServiceIDList, types.CoresCount),
	}
	return PVM.HostCallArgs{
		GeneralArgs: PVM.GeneralArgs{
			ServiceAccount:      &acct,
			ServiceID:           &sid,
			ServiceAccountState: &state,
			StorageKeyVal:       &storage,
		},
		AccumulateArgs: PVM.AccumulateArgs{
			ResultContextX: PVM.ResultContext{
				ServiceID:         sid,
				PartialState:      partial,
				DeferredTransfers: []types.DeferredTransfer{},
				ServiceBlobs:      map[types.OpaqueHash]types.ServiceBlob{},
				StorageKeyVal:     &storage,
			},
			ResultContextY: PVM.ResultContext{
				ServiceID:         sid,
				PartialState:      partial.DeepCopy(),
				DeferredTransfers: []types.DeferredTransfer{},
				ServiceBlobs:      map[types.OpaqueHash]types.ServiceBlob{},
				StorageKeyVal:     &storage,
			},
		},
		CodeHash: p.hash,
	}
}
