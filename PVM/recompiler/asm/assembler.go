package asm

// Assembler wraps a CodeBuffer and provides high-level instruction emit methods.
type Assembler struct {
	buf *CodeBuffer

	// exitTrampoline is the well-known per-block exit label: many emit helpers
	// jump to it without threading a handle through their signatures. Allocated
	// fresh on every Reset; bound by EmitExitTrampoline.
	exitTrampoline Label
}

func NewAssembler() *Assembler {
	a := &Assembler{buf: NewCodeBuffer()}
	a.exitTrampoline = a.buf.NewLabel()
	return a
}

// NewLabel allocates a fresh unbound label handle.
func (a *Assembler) NewLabel() Label { return a.buf.NewLabel() }

// ExitTrampoline returns the well-known exit-trampoline label for the current
// Reset cycle.
func (a *Assembler) ExitTrampoline() Label { return a.exitTrampoline }

// Finalize resolves all label fixups and returns the final machine code bytes.
func (a *Assembler) Finalize() ([]byte, error) {
	if err := a.buf.ResolveFixups(); err != nil {
		return nil, err
	}
	return a.buf.Bytes(), nil
}

// Len returns the current number of emitted bytes.
func (a *Assembler) Len() int { return a.buf.Len() }

// Buffer returns the underlying CodeBuffer for advanced usage.
func (a *Assembler) Buffer() *CodeBuffer { return a.buf }

// Reset clears all emitted code and labels.
func (a *Assembler) Reset() {
	a.buf.Reset()
	a.exitTrampoline = a.buf.NewLabel()
}
