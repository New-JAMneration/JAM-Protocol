// Package pvmcost defines the cost summary types that JIP-3 telemetry
// events 47 (Block executed) / 95 (Work-report guaranteed) / 101
// (Work-package validated) carry on the wire.
//
// These structs are PURE DATA — no Encode methods, no wire knowledge.
// Wire encoding lives in internal/telemetry (Encode...Cost free
// functions + Decoder methods). The split keeps PVM (the producer)
// from having to import telemetry (the consumer / encoder).
//
// Observability-only. Cost values are emitted to the JIP-3 aggregator
// for monitoring; they MUST NOT be embedded in consensus-serialized
// types (types.WorkReport, types.WorkResult, types.Guarantee, block
// hash inputs, conformance vectors). The CI guard in
// consensus_boundary_test.go enforces this by walking the consensus
// types and asserting none of them transitively contain a pvmcost
// type.
//
// Spec: https://github.com/polkadot-fellows/JIPs/blob/main/JIP-3.md
package pvmcost

// Gas is a u64 alias for clarity at field sites. Same shape as
// internal/types.Gas but lives here so pvmcost stays a leaf package.
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

// IsAuthorizedCost is the cost summary for the Is-Authorized PVM phase
// (JIP-3 event 95 component).
//
// Wire layout (40 bytes): Total ++ CompileNanos ++ HostCalls.
type IsAuthorizedCost struct {
	Total        ExecCost
	CompileNanos uint64
	HostCalls    ExecCost
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

// AccumulateCost is the cost summary for the Accumulate PVM phase
// (JIP-3 event 47 component).
//
// Wire layout (140 bytes):
//
//	u32 LE  AccumulateCalls
//	u32 LE  TransfersProcessed
//	u32 LE  ItemsAccumulated
//	16 B    Total                       (ExecCost)
//	u64 LE  CompileNanos
//	16 B    ReadWrite                   (ExecCost)
//	16 B    Lookup                      (ExecCost)
//	16 B    QuerySolicitForgetProvide   (ExecCost)
//	16 B    InfoNewUpgradeEject         (ExecCost)
//	16 B    Transfer                    (ExecCost)
//	u64 LE  TotalTransferGas            (Gas)
//	16 B    Other                       (ExecCost)
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
