package work_package

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/internal/pvmcost"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/stretchr/testify/require"
)

// stubCostPVM returns canned Psi_I / RefineInvoke results keyed by work-item
// index, so the test can plant distinct sentinel costs at the PVM leaf and
// assert they reach the WorkPackageTelemetryCost sidecar untouched (#974
// Phase 1 plumbing contract).
type stubCostPVM struct {
	isAuth PVM.Psi_I_ReturnType
	refine map[uint]PVM.RefineOutput
}

func (s *stubCostPVM) Psi_I(types.WorkPackage, types.CoreIndex, types.ByteSequence) PVM.Psi_I_ReturnType {
	return s.isAuth
}

func (s *stubCostPVM) RefineInvoke(in PVM.RefineInput) PVM.RefineOutput {
	return s.refine[in.WorkItemIndex]
}

// Distinct primes per field so a field-swap or a zero-out shows up.
func sentinelExec(seed uint64) pvmcost.ExecCost {
	return pvmcost.ExecCost{GasUsed: seed, ElapsedNanos: seed + 1}
}

func sentinelRefineCost(seed uint64) pvmcost.RefineCost {
	return pvmcost.RefineCost{
		Total:            sentinelExec(seed),
		CompileNanos:     seed + 10,
		HistoricalLookup: sentinelExec(seed + 20),
		MachineExpunge:   sentinelExec(seed + 30),
		PeekPokePages:    sentinelExec(seed + 40),
		Invoke:           sentinelExec(seed + 50),
		Other:            sentinelExec(seed + 60),
	}
}

func TestWorkReportCompute_CostSidecarReachesEmitLayer(t *testing.T) {
	isAuthCost := pvmcost.IsAuthorizedCost{
		Total:        sentinelExec(11),
		CompileNanos: 13,
		HostCalls:    sentinelExec(17),
	}
	refine0 := sentinelRefineCost(101)
	refine1 := sentinelRefineCost(211)

	// Item 1 reports ExportCount=1 but the stub returns no export segments,
	// forcing the BadExports failure branch: failed items must still keep
	// their cost slot so Refine stays index-aligned with Items.
	wp := &types.WorkPackage{
		Items: []types.WorkItem{
			{Service: 1, ExportCount: 0},
			{Service: 2, ExportCount: 1},
		},
	}

	stub := &stubCostPVM{
		isAuth: PVM.Psi_I_ReturnType{
			WorkExecResult: types.WorkExecResultOk,
			WorkOutput:     []byte("auth-ok"),
			Gas:            7,
			Cost:           isAuthCost,
		},
		refine: map[uint]PVM.RefineOutput{
			0: {WorkResult: types.WorkExecResultOk, RefineOutput: []byte{0x01}, ExportSegment: []types.ExportSegment{}, Gas: 5, Cost: refine0},
			1: {WorkResult: types.WorkExecResultOk, RefineOutput: []byte{0x02}, ExportSegment: []types.ExportSegment{}, Gas: 6, Cost: refine1},
		},
	}

	report, cost, err := WorkReportCompute(
		wp,
		types.CoreIndex(0),
		types.OpaqueHash{0xAA},
		types.ByteSequence("authorizer-code"),
		PVM.ExtrinsicDataMap{},
		nil,
		types.ServiceAccountState{},
		[]byte("work-package-bundle"),
		types.OpaqueHash{0xBB},
		stub,
	)
	require.NoError(t, err)

	// Sidecar carries the leaf sentinels untouched.
	require.Equal(t, isAuthCost, cost.IsAuthorized)
	require.Len(t, cost.Refine, len(wp.Items))
	require.Equal(t, refine0, cost.Refine[0])
	require.Equal(t, refine1, cost.Refine[1])

	// The failed item proves alignment: result 1 is BadExports yet its
	// cost slot is still the item-1 sentinel.
	require.Equal(t, types.WorkExecResultOk, report.Results[0].Result.Type)
	require.Equal(t, types.WorkExecResultType(types.WorkExecResultBadExports), report.Results[1].Result.Type)

	// Cost rides the sidecar only — WorkReport stays a pure consensus
	// type (the reflect guard in internal/pvmcost enforces this shape;
	// here we just confirm the report is intact and usable).
	require.Len(t, report.Results, 2)
	require.Equal(t, types.Gas(7), report.AuthGasUsed)
}

// Process's hand-off to Controller.TelemetryCost is asserted inside
// TestWorkPackageController_InitialProcess (work_package_controller_test.go),
// which already exercises the full Process path.
