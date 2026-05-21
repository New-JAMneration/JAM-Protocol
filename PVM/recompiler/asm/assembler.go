package asm

// Assembler wraps a CodeBuffer and provides high-level instruction emit methods.
type Assembler struct {
	buf *CodeBuffer
}

func NewAssembler() *Assembler {
	return &Assembler{buf: NewCodeBuffer()}
}

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
func (a *Assembler) Reset() { a.buf.Reset() }
