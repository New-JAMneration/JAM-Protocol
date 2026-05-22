package telemetry

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/pvmcost"
)

// Wire encoders + Decoder methods for the JIP-3 cost summary types.
// Types themselves live in internal/pvmcost (pure data, no wire
// knowledge) so PVM doesn't have to import telemetry. Per #775 Q2 +
// #974, callers may emit zero-filled costs until PVM instrumentation
// lands; a zero-value cost encodes to all-zero bytes (see *_test.go).

const (
	execCostEncodedSize         = 8 + 8
	isAuthorizedCostEncodedSize = execCostEncodedSize + 8 + execCostEncodedSize
	refineCostEncodedSize       = execCostEncodedSize + 8 + 5*execCostEncodedSize
	accumulateCostEncodedSize   = 4*3 + execCostEncodedSize + 8 + 5*execCostEncodedSize + 8 + execCostEncodedSize
)

// EncodeExecCost produces the wire bytes for c.
func EncodeExecCost(c pvmcost.ExecCost) []byte {
	out := make([]byte, 0, execCostEncodedSize)
	out = append(out, EncodeU64(c.GasUsed)...)
	out = append(out, EncodeU64(c.ElapsedNanos)...)
	return out
}

// ReadExecCost reads an ExecCost from the decoder.
func (d *Decoder) ReadExecCost() (pvmcost.ExecCost, error) {
	var c pvmcost.ExecCost
	var err error
	if c.GasUsed, err = d.ReadU64(); err != nil {
		return pvmcost.ExecCost{}, fmt.Errorf("ExecCost.GasUsed: %w", err)
	}
	if c.ElapsedNanos, err = d.ReadU64(); err != nil {
		return pvmcost.ExecCost{}, fmt.Errorf("ExecCost.ElapsedNanos: %w", err)
	}
	return c, nil
}

// EncodeIsAuthorizedCost produces the wire bytes for c.
func EncodeIsAuthorizedCost(c pvmcost.IsAuthorizedCost) []byte {
	out := make([]byte, 0, isAuthorizedCostEncodedSize)
	out = append(out, EncodeExecCost(c.Total)...)
	out = append(out, EncodeU64(c.CompileNanos)...)
	out = append(out, EncodeExecCost(c.HostCalls)...)
	return out
}

// ReadIsAuthorizedCost reads an IsAuthorizedCost from the decoder.
func (d *Decoder) ReadIsAuthorizedCost() (pvmcost.IsAuthorizedCost, error) {
	var c pvmcost.IsAuthorizedCost
	var err error
	if c.Total, err = d.ReadExecCost(); err != nil {
		return pvmcost.IsAuthorizedCost{}, fmt.Errorf("IsAuthorizedCost.Total: %w", err)
	}
	if c.CompileNanos, err = d.ReadU64(); err != nil {
		return pvmcost.IsAuthorizedCost{}, fmt.Errorf("IsAuthorizedCost.CompileNanos: %w", err)
	}
	if c.HostCalls, err = d.ReadExecCost(); err != nil {
		return pvmcost.IsAuthorizedCost{}, fmt.Errorf("IsAuthorizedCost.HostCalls: %w", err)
	}
	return c, nil
}

// EncodeRefineCost produces the wire bytes for c.
func EncodeRefineCost(c pvmcost.RefineCost) []byte {
	out := make([]byte, 0, refineCostEncodedSize)
	out = append(out, EncodeExecCost(c.Total)...)
	out = append(out, EncodeU64(c.CompileNanos)...)
	out = append(out, EncodeExecCost(c.HistoricalLookup)...)
	out = append(out, EncodeExecCost(c.MachineExpunge)...)
	out = append(out, EncodeExecCost(c.PeekPokePages)...)
	out = append(out, EncodeExecCost(c.Invoke)...)
	out = append(out, EncodeExecCost(c.Other)...)
	return out
}

// ReadRefineCost reads a RefineCost from the decoder.
func (d *Decoder) ReadRefineCost() (pvmcost.RefineCost, error) {
	var c pvmcost.RefineCost
	var err error
	if c.Total, err = d.ReadExecCost(); err != nil {
		return pvmcost.RefineCost{}, fmt.Errorf("RefineCost.Total: %w", err)
	}
	if c.CompileNanos, err = d.ReadU64(); err != nil {
		return pvmcost.RefineCost{}, fmt.Errorf("RefineCost.CompileNanos: %w", err)
	}
	if c.HistoricalLookup, err = d.ReadExecCost(); err != nil {
		return pvmcost.RefineCost{}, fmt.Errorf("RefineCost.HistoricalLookup: %w", err)
	}
	if c.MachineExpunge, err = d.ReadExecCost(); err != nil {
		return pvmcost.RefineCost{}, fmt.Errorf("RefineCost.MachineExpunge: %w", err)
	}
	if c.PeekPokePages, err = d.ReadExecCost(); err != nil {
		return pvmcost.RefineCost{}, fmt.Errorf("RefineCost.PeekPokePages: %w", err)
	}
	if c.Invoke, err = d.ReadExecCost(); err != nil {
		return pvmcost.RefineCost{}, fmt.Errorf("RefineCost.Invoke: %w", err)
	}
	if c.Other, err = d.ReadExecCost(); err != nil {
		return pvmcost.RefineCost{}, fmt.Errorf("RefineCost.Other: %w", err)
	}
	return c, nil
}

// EncodeAccumulateCost produces the wire bytes for c.
func EncodeAccumulateCost(c pvmcost.AccumulateCost) []byte {
	out := make([]byte, 0, accumulateCostEncodedSize)
	out = append(out, EncodeU32(c.AccumulateCalls)...)
	out = append(out, EncodeU32(c.TransfersProcessed)...)
	out = append(out, EncodeU32(c.ItemsAccumulated)...)
	out = append(out, EncodeExecCost(c.Total)...)
	out = append(out, EncodeU64(c.CompileNanos)...)
	out = append(out, EncodeExecCost(c.ReadWrite)...)
	out = append(out, EncodeExecCost(c.Lookup)...)
	out = append(out, EncodeExecCost(c.QuerySolicitForgetProvide)...)
	out = append(out, EncodeExecCost(c.InfoNewUpgradeEject)...)
	out = append(out, EncodeExecCost(c.Transfer)...)
	out = append(out, EncodeU64(c.TotalTransferGas)...)
	out = append(out, EncodeExecCost(c.Other)...)
	return out
}

// ReadAccumulateCost reads an AccumulateCost from the decoder.
func (d *Decoder) ReadAccumulateCost() (pvmcost.AccumulateCost, error) {
	var c pvmcost.AccumulateCost
	var err error
	if c.AccumulateCalls, err = d.ReadU32(); err != nil {
		return pvmcost.AccumulateCost{}, fmt.Errorf("AccumulateCost.AccumulateCalls: %w", err)
	}
	if c.TransfersProcessed, err = d.ReadU32(); err != nil {
		return pvmcost.AccumulateCost{}, fmt.Errorf("AccumulateCost.TransfersProcessed: %w", err)
	}
	if c.ItemsAccumulated, err = d.ReadU32(); err != nil {
		return pvmcost.AccumulateCost{}, fmt.Errorf("AccumulateCost.ItemsAccumulated: %w", err)
	}
	if c.Total, err = d.ReadExecCost(); err != nil {
		return pvmcost.AccumulateCost{}, fmt.Errorf("AccumulateCost.Total: %w", err)
	}
	if c.CompileNanos, err = d.ReadU64(); err != nil {
		return pvmcost.AccumulateCost{}, fmt.Errorf("AccumulateCost.CompileNanos: %w", err)
	}
	if c.ReadWrite, err = d.ReadExecCost(); err != nil {
		return pvmcost.AccumulateCost{}, fmt.Errorf("AccumulateCost.ReadWrite: %w", err)
	}
	if c.Lookup, err = d.ReadExecCost(); err != nil {
		return pvmcost.AccumulateCost{}, fmt.Errorf("AccumulateCost.Lookup: %w", err)
	}
	if c.QuerySolicitForgetProvide, err = d.ReadExecCost(); err != nil {
		return pvmcost.AccumulateCost{}, fmt.Errorf("AccumulateCost.QuerySolicitForgetProvide: %w", err)
	}
	if c.InfoNewUpgradeEject, err = d.ReadExecCost(); err != nil {
		return pvmcost.AccumulateCost{}, fmt.Errorf("AccumulateCost.InfoNewUpgradeEject: %w", err)
	}
	if c.Transfer, err = d.ReadExecCost(); err != nil {
		return pvmcost.AccumulateCost{}, fmt.Errorf("AccumulateCost.Transfer: %w", err)
	}
	if c.TotalTransferGas, err = d.ReadU64(); err != nil {
		return pvmcost.AccumulateCost{}, fmt.Errorf("AccumulateCost.TotalTransferGas: %w", err)
	}
	if c.Other, err = d.ReadExecCost(); err != nil {
		return pvmcost.AccumulateCost{}, fmt.Errorf("AccumulateCost.Other: %w", err)
	}
	return c, nil
}
