package accumulation

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

type OuterAccumulationInput struct {
	GasLimit                     types.Gas                 // g    gas-limit
	WorkReports                  []types.WorkReport        // w    a sequence of work-reports
	InitPartialStateSet          types.PartialStateSet     // o    an initial partial-state
	ServicesWithFreeAccumulation types.AlwaysAccumulateMap // f    a dictionary of services enjoying free accumulation
}

type OuterAccumulationOutput struct {
	NumberOfWorkResultsAccumulated types.U64
	PartialStateSet                types.PartialStateSet    // a posterior state-context
	DeferredTransfers              []types.DeferredTransfer // resultant deferred transfers
	ServiceHashSet                 types.ServiceHashSet     // service/hash pairs
	ServiceGasUsedList             types.ServiceGasUsedList // service/gas pairs
}

type ParallelizedAccumulationInput struct {
	PartialStateSet     types.PartialStateSet     // o   initial state-context
	WorkReports         []types.WorkReport        // w   a sequence of work-reports
	AlwaysAccumulateMap types.AlwaysAccumulateMap // f   a dictionary of privileged always-accumulate services
}

type ParallelizedAccumulationOutput struct {
	PartialStateSet    types.PartialStateSet    // a posterior state-context
	DeferredTransfers  []types.DeferredTransfer // resultant deferred transfers
	ServiceHashSet     types.ServiceHashSet     // service/hash pairs
	ServiceGasUsedList types.ServiceGasUsedList // service/gas pairs
}

type SingleServiceAccumulationInput struct {
	PartialStateSet     types.PartialStateSet     // o   initial state-context
	WorkReports         []types.WorkReport        // w   a sequence of work-reports
	AlwaysAccumulateMap types.AlwaysAccumulateMap // f   a dictionary of privileged always-accumulate services
	ServiceId           types.ServiceId           // s   a service index
}

type SingleServiceAccumulationOutput struct {
	PartialStateSet    types.PartialStateSet    // an alterations state-context
	DeferredTransfers  []types.DeferredTransfer // a sequence of transfers
	AccumulationOutput *types.OpaqueHash        // a possible accumulation-output
	GasUsed            types.Gas                // the actual PVM gas used
}
