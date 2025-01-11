package store

import (
	"sync"

	types "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type BeefyCommitmentOutputs struct {
	mu sync.RWMutex
	C  *types.BeefyCommitmentOutput
}

func NewBeefyCommitmentOutput() *BeefyCommitmentOutputs {
	return &BeefyCommitmentOutputs{
		C: &types.BeefyCommitmentOutput{},
	}
}

func (b *BeefyCommitmentOutputs) GetBeefyCommitmentOutput() types.BeefyCommitmentOutput {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return *b.C
}

func (b *BeefyCommitmentOutputs) SetBeefyCommitmentOutput(c types.BeefyCommitmentOutput) {
	b.mu.Lock()
	defer b.mu.Unlock()
	*b.C = append(*b.C, c...)
}
