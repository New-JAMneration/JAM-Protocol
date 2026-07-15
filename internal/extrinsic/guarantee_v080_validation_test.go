package extrinsic

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	ReportsErrorCode "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/reports"
)

func requireReportsErrorCode(t *testing.T, err error, want types.ErrorCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected reports error code %d, got nil", want)
	}
	got, ok := err.(*types.ErrorCode)
	if !ok {
		t.Fatalf("expected *types.ErrorCode, got %T: %v", err, err)
	}
	if *got != want {
		t.Fatalf("reports error code = %d, want %d", *got, want)
	}
}

func TestReportGuaranteeValidateRejectsCredentialOutsideTwoToThree(t *testing.T) {
	guarantee := types.ReportGuarantee{
		Report: types.WorkReport{
			Results: []types.WorkResult{{}},
		},
		Signatures: make([]types.ValidatorSignature, types.GuaranteeMaxCount+1),
	}

	if err := guarantee.Validate(); err == nil || err.Error() != "insufficient_guarantees" {
		t.Fatalf("Validate() error = %v, want insufficient_guarantees", err)
	}
}

func TestGuaranteeControllerValidateWorkReportsChecksErasureShards(t *testing.T) {
	blockchain.ResetInstance()
	t.Cleanup(blockchain.ResetInstance)
	cs := blockchain.GetInstance()

	const core = types.CoreIndex(0)
	authorizer := types.AuthorizerHash{0xAA}
	alpha := make(types.AuthPools, types.CoresCount)
	alpha[core] = types.AuthPool{authorizer}
	cs.GetPriorStates().SetAlpha(alpha)
	cs.GetIntermediateStates().SetRhoDoubleDagger(make(types.AvailabilityAssignments, types.CoresCount))
	cs.GetPosteriorStates().SetKappa(make(types.ValidatorsData, 3))

	controller := NewGuaranteeController()
	controller.Set([]types.ReportGuarantee{{
		Report: types.WorkReport{
			PackageSpec:    types.WorkPackageSpec{ErasureShards: 2},
			CoreIndex:      core,
			AuthorizerHash: types.OpaqueHash(authorizer),
		},
	}})

	requireReportsErrorCode(t, controller.ValidateWorkReports(), ReportsErrorCode.BadErasureShards)

	controller.Guarantees[0].Report.PackageSpec.ErasureShards = 3
	if err := controller.ValidateWorkReports(); err != nil {
		t.Fatalf("ValidateWorkReports() with matching shard count: %v", err)
	}
}

func TestGuaranteeControllerValidateContextsV080Fields(t *testing.T) {
	anchor := types.HeaderHash{0x11}
	anchorStateRoot := types.StateRoot{0x22}
	anchorBeefyRoot := types.BeefyRoot{0x33}
	lookupAnchor := types.HeaderHash{0x44}
	lookupStateRoot := types.StateRoot{0x55}
	const anchorSlot = types.TimeSlot(80)
	const lookupSlot = types.TimeSlot(90)

	baseContext := types.RefineContext{
		Anchor:                anchor,
		AnchorSlot:            anchorSlot,
		StateRoot:             anchorStateRoot,
		BeefyRoot:             anchorBeefyRoot,
		LookupAnchor:          lookupAnchor,
		LookupAnchorSlot:      lookupSlot,
		LookupAnchorStateRoot: lookupStateRoot,
	}

	tests := []struct {
		name          string
		mutate        func(*types.RefineContext)
		includeLookup bool
		wantErrorCode *types.ErrorCode
	}{
		{name: "valid", includeLookup: true},
		{
			name:          "anchor timeslot mismatch",
			includeLookup: true,
			mutate: func(c *types.RefineContext) {
				c.AnchorSlot++
			},
			wantErrorCode: ptrErrorCode(ReportsErrorCode.AnchorNotRecent),
		},
		{
			name:          "lookup state root mismatch",
			includeLookup: true,
			mutate: func(c *types.RefineContext) {
				c.LookupAnchorStateRoot[0] ^= 0xFF
			},
			wantErrorCode: ptrErrorCode(ReportsErrorCode.BadStateRoot),
		},
		{
			name:          "lookup anchor missing",
			includeLookup: false,
			wantErrorCode: ptrErrorCode(ReportsErrorCode.LookupAnchorNotRecent),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			blockchain.ResetInstance()
			t.Cleanup(blockchain.ResetInstance)
			cs := blockchain.GetInstance()
			cs.AddBlock(types.Block{Header: types.Header{Slot: 100}})
			cs.GetIntermediateStates().SetBetaHDagger(types.BlocksHistory{{
				HeaderHash: anchor,
				StateRoot:  anchorStateRoot,
				BeefyRoot:  types.OpaqueHash(anchorBeefyRoot),
				Timeslot:   anchorSlot,
			}})
			if tc.includeLookup {
				cs.AppendAncestry(types.Ancestry{{
					Slot:       lookupSlot,
					HeaderHash: lookupAnchor,
					StateRoot:  lookupStateRoot,
				}})
			}

			context := baseContext
			if tc.mutate != nil {
				tc.mutate(&context)
			}
			controller := NewGuaranteeController()
			controller.Set([]types.ReportGuarantee{{
				Report: types.WorkReport{Context: context},
			}})

			err := controller.ValidateContexts()
			if tc.wantErrorCode == nil {
				if err != nil {
					t.Fatalf("ValidateContexts(): %v", err)
				}
				return
			}
			requireReportsErrorCode(t, err, *tc.wantErrorCode)
		})
	}
}

func ptrErrorCode(code types.ErrorCode) *types.ErrorCode {
	return &code
}
