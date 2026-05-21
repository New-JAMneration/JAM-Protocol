package PVM

// InlineSnapshotPtrs returns pointers to the VMState's inline buffers
// (registersBuf / gasBuf). Callers — typically the JIT host-call loop —
// fill these buffers in-place (e.g. via JITContext.ReadRegistersInto)
// and then call BindInlineSnapshot to publish them through the public
// Registers / Gas pointer fields.
//
// Because the buffers live inside the VMState struct itself, a
// stack-allocated VMState keeps the whole snapshot heap-free: Omega
// functions mutate the buffers directly via vm.Registers / vm.Gas, and
// the caller reads them back through InlineSnapshotValues after Omega
// returns to commit changes to its own source of truth (e.g. the
// JIT control region).
//
// Interpreter does not use this path: it aliases Registers / Gas at
// its own live fields, so mutations propagate without any snapshot.
func (v *VMState) InlineSnapshotPtrs() (*Registers, *Gas) {
	return &v.registersBuf, &v.gasBuf
}

// BindInlineSnapshot publishes the inline buffers through the public
// Registers / Gas pointer fields. Call this after filling the inline
// buffers (via InlineSnapshotPtrs or BindInlineSnapshotValues) so that
// Omega functions observe them through the regular VMState API.
func (v *VMState) BindInlineSnapshot() {
	v.Registers = &v.registersBuf
	v.Gas = &v.gasBuf
}

// BindInlineSnapshotValues is a convenience for callers that already
// hold Registers / Gas by value: it copies them into the inline
// buffers and then publishes through Registers / Gas pointers. Useful
// for tests or for non-JIT callers that cannot use the zero-copy
// InlineSnapshotPtrs path.
func (v *VMState) BindInlineSnapshotValues(regs Registers, gas Gas) {
	v.registersBuf = regs
	v.gasBuf = gas
	v.BindInlineSnapshot()
}

// InlineSnapshotValues returns the current value of the inline buffers.
// Valid after Omega execution; callers use this to write back to their
// source of truth (e.g. the JIT control region).
func (v *VMState) InlineSnapshotValues() (Registers, Gas) {
	return v.registersBuf, v.gasBuf
}
