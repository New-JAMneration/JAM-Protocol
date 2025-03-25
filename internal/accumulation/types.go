package accumulation

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

type OuterAccumulationInput struct {
	GasLimit                     types.Gas                 // gas-limit
	WorkReports                  []types.WorkReport        // a sequence of work-reports
	InitPartialStateSet          types.PartialStateSet     // an initial partial-state
	ServicesWithFreeAccumulation types.AlwaysAccumulateMap // a dictionary of services enjoying free accumulation
}

type OuterAccumulationOutput struct {
	NumberOfWorkResultsAccumulated types.U64
	PartialStateSet                types.PartialStateSet    // a posterior state-context
	DeferredTransfers              []types.DeferredTransfer // resultant deferred transfers
	ServiceHashSet                 types.ServiceHashSet     // service/hash pairs
	ServiceGasUsedList             types.ServiceGasUsedList // service/gas pairs
}

type ParallelizedAccumulationInput struct {
	PartialStateSet     types.PartialStateSet     // initial state-context
	WorkReports         []types.WorkReport        // a sequence of work-reports
	AlwaysAccumulateMap types.AlwaysAccumulateMap // a dictionary of privileged always-accumulate services
}

type ParallelizedAccumulationOutput struct {
	PartialStateSet    types.PartialStateSet    // a posterior state-context
	DeferredTransfers  []types.DeferredTransfer // resultant deferred transfers
	ServiceHashSet     types.ServiceHashSet     // service/hash pairs
	ServiceGasUsedList types.ServiceGasUsedList // service/gas pairs
}

type SingleServiceAccumulationInput struct {
	PartialStateSet     types.PartialStateSet     // initial state-context
	WorkReports         []types.WorkReport        // a sequence of work-reports
	AlwaysAccumulateMap types.AlwaysAccumulateMap // a dictionary of privileged always-accumulate services
	ServiceId           types.ServiceId           // a service index
}

type SingleServiceAccumulationOutput struct {
	PartialStateSet    types.PartialStateSet    // an alterations state-context
	DeferredTransfers  []types.DeferredTransfer // a sequence of transfers
	AccumulationOutput *types.OpaqueHash        // a possible accumulation-output
	GasUsed            types.Gas                // the actual PVM gas used
}
