//go:build linux && amd64

package recompiler

import (
	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
)

// TODO: this is a draft, will be updated until all other components are implemented.

// CompiledBlock holds metadata for a single compiled PVM basic block.
type CompiledBlock struct {
	PVMStartPC   PVM.ProgramCounter // first PVM instruction PC in this block
	PVMEndPC     PVM.ProgramCounter // one-past-last PVM instruction PC
	NativeAddr   uintptr            // callable native code address (precomputed)
	NativeOffset int                // byte offset into ExecutableMemory (for debug)
	NativeSize   int                // size of emitted native code in bytes
	GasCost      int64              // total gas cost for this block (= number of PVM instructions)
	InstrCount   int                // number of PVM instructions in this block
}

// CodeCache maps PVM Program Counters to compiled basic blocks.
// Not thread-safe: each PVM instance runs on a single goroutine.
type CodeCache struct {
	blocks map[PVM.ProgramCounter]*CompiledBlock
	em     *ExecutableMemory // optional; reset on Invalidate when writable
}

// NewCodeCache creates an empty CodeCache.
func NewCodeCache() *CodeCache {
	return &CodeCache{
		blocks: make(map[PVM.ProgramCounter]*CompiledBlock),
	}
}

// Get looks up a compiled block by its PVM start PC.
// Returns nil if the block has not been compiled.
func (cc *CodeCache) Get(pc PVM.ProgramCounter) *CompiledBlock {
	return cc.blocks[pc]
}

// Put stores a compiled block in the cache, keyed by its PVMStartPC.
func (cc *CodeCache) Put(block *CompiledBlock) {
	cc.blocks[block.PVMStartPC] = block
}

// Has reports whether a block has been compiled for the given PC.
func (cc *CodeCache) Has(pc PVM.ProgramCounter) bool {
	_, ok := cc.blocks[pc]
	return ok
}

// BindExecutableMemory associates the JIT code arena with this cache so
// Invalidate can reclaim emitted native code.
func (cc *CodeCache) BindExecutableMemory(em *ExecutableMemory) {
	cc.em = em
}

// Invalidate discards all compiled blocks and resets bound ExecutableMemory
// when it is writable (INT3-fill and used=0).
func (cc *CodeCache) Invalidate() error {
	clear(cc.blocks)
	if cc.em == nil {
		return nil
	}
	return cc.em.Reset()
}
