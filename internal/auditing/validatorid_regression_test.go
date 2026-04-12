package auditing

import (
	"crypto/ed25519"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Regression (#935): each AuditReport built by ComputeInitialAuditAssignment
// must carry the originating ValidatorID, not default to 0.
func TestInitialAssignment_SetsValidatorID(t *testing.T) {
	t.Helper()

	types.SetTinyMode()
	blockchain.ResetInstance()
	t.Cleanup(blockchain.ResetInstance)

	cs := blockchain.GetInstance()
	var entropy types.BandersnatchVrfSignature
	entropy[0] = 1
	cs.GetProcessingBlockPointer().SetHeader(types.Header{
		Slot:          100,
		AuthorIndex:   0,
		EntropySource: entropy,
	})

	var h0, h1 types.WorkPackageHash
	h0[0] = 1
	h1[0] = 2

	report0 := types.WorkReport{
		PackageSpec: types.WorkPackageSpec{Hash: h0},
		CoreIndex:   0,
		Results:     []types.WorkResult{{}},
	}
	report1 := types.WorkReport{
		PackageSpec: types.WorkPackageSpec{Hash: h1},
		CoreIndex:   1,
		Results:     []types.WorkResult{{}},
	}

	Q := make([]*types.WorkReport, types.CoresCount)
	Q[0] = &report0
	Q[1] = &report1

	seed := make([]byte, ed25519.SeedSize)
	seed[0] = 42
	_ = ed25519.NewKeyFromSeed(seed)

	validatorIndex := types.ValidatorIndex(4)
	got := buildInitialAuditAssignmentFromCoreOrder(Q, validatorIndex, []types.U32{1, 0})

	require.Len(t, got, 2)
	for _, audit := range got {
		assert.Equal(t, validatorIndex, audit.ValidatorID, "ValidatorID must be set, not default 0")
		assert.False(t, audit.AuditResult)
	}
}
