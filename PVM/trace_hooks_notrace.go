//go:build !trace

package PVM

func (interp *Interpreter) noteTraceMemAfterSuccessfulStore(memIndex uint32, immediate uint64) {}

func (interp *Interpreter) noteTraceMemAfterSuccessfulLoad(vX uint32, memVal uint64) {}

func (interp *Interpreter) recordInstrTraceStepAfterMeta(instr *InstrMeta, src1Val, src2Val uint64) {}
