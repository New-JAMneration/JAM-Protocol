//go:build trace

package PVM

func (interp *Interpreter) noteTraceMemAfterSuccessfulStore(memIndex uint32, immediate uint64) {
	if interp == nil || interp.Trace == nil {
		return
	}
	interp.LastStore.Addr = memIndex
	interp.LastStore.Val = immediate
	interp.LastStore.Active = true
}

func (interp *Interpreter) noteTraceMemAfterSuccessfulLoad(vX uint32, memVal uint64) {
	if interp == nil || interp.Trace == nil {
		return
	}
	interp.LastLoad.Addr = vX
	interp.LastLoad.Val = memVal
	interp.LastLoad.Active = true
}

func (interp *Interpreter) recordInstrTraceStepAfterMeta(instr *InstrMeta, src1Val, src2Val uint64) {
	if interp == nil || interp.Trace == nil {
		return
	}
	var dstVal uint64
	if instr.Dst != 0xff {
		dstVal = interp.Registers[instr.Dst]
	}
	var loadAddr, storeAddr uint32
	var loadVal, storeVal uint64
	if interp.LastLoad.Active {
		loadAddr = interp.LastLoad.Addr
		loadVal = interp.LastLoad.Val
	}
	if interp.LastStore.Active {
		storeAddr = interp.LastStore.Addr
		storeVal = interp.LastStore.Val
	}
	interp.Trace.RecordStep(
		uint32(instr.PC), instr.Opcode,
		instr.Dst, instr.Src[0], instr.Src[1],
		dstVal, src1Val, src2Val,
		int64(interp.Gas),
		loadAddr, loadVal,
		storeAddr, storeVal,
	)
	interp.LastLoad.Active = false
	interp.LastStore.Active = false
}
