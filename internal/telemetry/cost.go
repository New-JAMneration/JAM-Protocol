package telemetry

import "fmt"

// Cost types used by JIP-3 events 47 (Block executed), 95 (Work-report
// guaranteed), and 101 (Work-package validated). Per #775 Q2, callers
// emit zero-filled Costs until PVM cost instrumentation lands; the
// types are wired here so emitters don't carry placeholder logic.

// Gas is a u64 alias for clarity at field sites.
type Gas = uint64

// ExecCost summarises a chunk of PVM execution.
//
// Wire layout (16 bytes):
//
//	u64 LE  GasUsed       (Gas)
//	u64 LE  ElapsedNanos  (wall-clock time in ns)
type ExecCost struct {
	GasUsed      Gas
	ElapsedNanos uint64
}

const execCostEncodedSize = 8 + 8

// Encode produces the wire bytes for c.
func (c ExecCost) Encode() []byte {
	out := make([]byte, 0, execCostEncodedSize)
	out = append(out, EncodeU64(c.GasUsed)...)
	out = append(out, EncodeU64(c.ElapsedNanos)...)
	return out
}

// ReadExecCost reads an ExecCost from the decoder.
func (d *Decoder) ReadExecCost() (ExecCost, error) {
	var c ExecCost
	var err error
	if c.GasUsed, err = d.ReadU64(); err != nil {
		return ExecCost{}, fmt.Errorf("ExecCost.GasUsed: %w", err)
	}
	if c.ElapsedNanos, err = d.ReadU64(); err != nil {
		return ExecCost{}, fmt.Errorf("ExecCost.ElapsedNanos: %w", err)
	}
	return c, nil
}

// IsAuthorizedCost is the cost summary for the Is-Authorized PVM phase
// (JIP-3 event 95 component).
//
// Wire layout (40 bytes): Total ++ CompileNanos ++ HostCalls.
type IsAuthorizedCost struct {
	Total        ExecCost
	CompileNanos uint64
	HostCalls    ExecCost
}

const isAuthorizedCostEncodedSize = execCostEncodedSize + 8 + execCostEncodedSize

// Encode produces the wire bytes for c.
func (c IsAuthorizedCost) Encode() []byte {
	out := make([]byte, 0, isAuthorizedCostEncodedSize)
	out = append(out, c.Total.Encode()...)
	out = append(out, EncodeU64(c.CompileNanos)...)
	out = append(out, c.HostCalls.Encode()...)
	return out
}

// ReadIsAuthorizedCost reads an IsAuthorizedCost from the decoder.
func (d *Decoder) ReadIsAuthorizedCost() (IsAuthorizedCost, error) {
	var c IsAuthorizedCost
	var err error
	if c.Total, err = d.ReadExecCost(); err != nil {
		return IsAuthorizedCost{}, fmt.Errorf("IsAuthorizedCost.Total: %w", err)
	}
	if c.CompileNanos, err = d.ReadU64(); err != nil {
		return IsAuthorizedCost{}, fmt.Errorf("IsAuthorizedCost.CompileNanos: %w", err)
	}
	if c.HostCalls, err = d.ReadExecCost(); err != nil {
		return IsAuthorizedCost{}, fmt.Errorf("IsAuthorizedCost.HostCalls: %w", err)
	}
	return c, nil
}

// RefineCost is the cost summary for the Refine PVM phase (JIP-3 event
// 101 component).
//
// Wire layout (104 bytes): Total ++ CompileNanos ++ five ExecCosts for
// host-call buckets ++ Other.
type RefineCost struct {
	Total            ExecCost
	CompileNanos     uint64
	HistoricalLookup ExecCost
	MachineExpunge   ExecCost
	PeekPokePages    ExecCost
	Invoke           ExecCost
	Other            ExecCost
}

const refineCostEncodedSize = execCostEncodedSize + 8 + 5*execCostEncodedSize

// Encode produces the wire bytes for c.
func (c RefineCost) Encode() []byte {
	out := make([]byte, 0, refineCostEncodedSize)
	out = append(out, c.Total.Encode()...)
	out = append(out, EncodeU64(c.CompileNanos)...)
	out = append(out, c.HistoricalLookup.Encode()...)
	out = append(out, c.MachineExpunge.Encode()...)
	out = append(out, c.PeekPokePages.Encode()...)
	out = append(out, c.Invoke.Encode()...)
	out = append(out, c.Other.Encode()...)
	return out
}

// ReadRefineCost reads a RefineCost from the decoder.
func (d *Decoder) ReadRefineCost() (RefineCost, error) {
	var c RefineCost
	var err error
	if c.Total, err = d.ReadExecCost(); err != nil {
		return RefineCost{}, fmt.Errorf("RefineCost.Total: %w", err)
	}
	if c.CompileNanos, err = d.ReadU64(); err != nil {
		return RefineCost{}, fmt.Errorf("RefineCost.CompileNanos: %w", err)
	}
	if c.HistoricalLookup, err = d.ReadExecCost(); err != nil {
		return RefineCost{}, fmt.Errorf("RefineCost.HistoricalLookup: %w", err)
	}
	if c.MachineExpunge, err = d.ReadExecCost(); err != nil {
		return RefineCost{}, fmt.Errorf("RefineCost.MachineExpunge: %w", err)
	}
	if c.PeekPokePages, err = d.ReadExecCost(); err != nil {
		return RefineCost{}, fmt.Errorf("RefineCost.PeekPokePages: %w", err)
	}
	if c.Invoke, err = d.ReadExecCost(); err != nil {
		return RefineCost{}, fmt.Errorf("RefineCost.Invoke: %w", err)
	}
	if c.Other, err = d.ReadExecCost(); err != nil {
		return RefineCost{}, fmt.Errorf("RefineCost.Other: %w", err)
	}
	return c, nil
}

// AccumulateCost is the cost summary for the Accumulate PVM phase
// (JIP-3 event 47 component).
//
// Wire layout (140 bytes):
//
//	u32 LE  AccumulateCalls
//	u32 LE  TransfersProcessed
//	u32 LE  ItemsAccumulated
//	16 B    Total            (ExecCost)
//	u64 LE  CompileNanos
//	16 B    ReadWrite        (ExecCost)
//	16 B    Lookup           (ExecCost)
//	16 B    QuerySolicitForgetProvide (ExecCost)
//	16 B    InfoNewUpgradeEject       (ExecCost)
//	16 B    Transfer         (ExecCost)
//	u64 LE  TotalTransferGas (Gas)
//	16 B    Other            (ExecCost)
type AccumulateCost struct {
	AccumulateCalls           uint32
	TransfersProcessed        uint32
	ItemsAccumulated          uint32
	Total                     ExecCost
	CompileNanos              uint64
	ReadWrite                 ExecCost
	Lookup                    ExecCost
	QuerySolicitForgetProvide ExecCost
	InfoNewUpgradeEject       ExecCost
	Transfer                  ExecCost
	TotalTransferGas          Gas
	Other                     ExecCost
}

const accumulateCostEncodedSize = 4*3 + execCostEncodedSize + 8 + 5*execCostEncodedSize + 8 + execCostEncodedSize

// Encode produces the wire bytes for c.
func (c AccumulateCost) Encode() []byte {
	out := make([]byte, 0, accumulateCostEncodedSize)
	out = append(out, EncodeU32(c.AccumulateCalls)...)
	out = append(out, EncodeU32(c.TransfersProcessed)...)
	out = append(out, EncodeU32(c.ItemsAccumulated)...)
	out = append(out, c.Total.Encode()...)
	out = append(out, EncodeU64(c.CompileNanos)...)
	out = append(out, c.ReadWrite.Encode()...)
	out = append(out, c.Lookup.Encode()...)
	out = append(out, c.QuerySolicitForgetProvide.Encode()...)
	out = append(out, c.InfoNewUpgradeEject.Encode()...)
	out = append(out, c.Transfer.Encode()...)
	out = append(out, EncodeU64(c.TotalTransferGas)...)
	out = append(out, c.Other.Encode()...)
	return out
}

// ReadAccumulateCost reads an AccumulateCost from the decoder.
func (d *Decoder) ReadAccumulateCost() (AccumulateCost, error) {
	var c AccumulateCost
	var err error
	if c.AccumulateCalls, err = d.ReadU32(); err != nil {
		return AccumulateCost{}, fmt.Errorf("AccumulateCost.AccumulateCalls: %w", err)
	}
	if c.TransfersProcessed, err = d.ReadU32(); err != nil {
		return AccumulateCost{}, fmt.Errorf("AccumulateCost.TransfersProcessed: %w", err)
	}
	if c.ItemsAccumulated, err = d.ReadU32(); err != nil {
		return AccumulateCost{}, fmt.Errorf("AccumulateCost.ItemsAccumulated: %w", err)
	}
	if c.Total, err = d.ReadExecCost(); err != nil {
		return AccumulateCost{}, fmt.Errorf("AccumulateCost.Total: %w", err)
	}
	if c.CompileNanos, err = d.ReadU64(); err != nil {
		return AccumulateCost{}, fmt.Errorf("AccumulateCost.CompileNanos: %w", err)
	}
	if c.ReadWrite, err = d.ReadExecCost(); err != nil {
		return AccumulateCost{}, fmt.Errorf("AccumulateCost.ReadWrite: %w", err)
	}
	if c.Lookup, err = d.ReadExecCost(); err != nil {
		return AccumulateCost{}, fmt.Errorf("AccumulateCost.Lookup: %w", err)
	}
	if c.QuerySolicitForgetProvide, err = d.ReadExecCost(); err != nil {
		return AccumulateCost{}, fmt.Errorf("AccumulateCost.QuerySolicitForgetProvide: %w", err)
	}
	if c.InfoNewUpgradeEject, err = d.ReadExecCost(); err != nil {
		return AccumulateCost{}, fmt.Errorf("AccumulateCost.InfoNewUpgradeEject: %w", err)
	}
	if c.Transfer, err = d.ReadExecCost(); err != nil {
		return AccumulateCost{}, fmt.Errorf("AccumulateCost.Transfer: %w", err)
	}
	if c.TotalTransferGas, err = d.ReadU64(); err != nil {
		return AccumulateCost{}, fmt.Errorf("AccumulateCost.TotalTransferGas: %w", err)
	}
	if c.Other, err = d.ReadExecCost(); err != nil {
		return AccumulateCost{}, fmt.Errorf("AccumulateCost.Other: %w", err)
	}
	return c, nil
}
