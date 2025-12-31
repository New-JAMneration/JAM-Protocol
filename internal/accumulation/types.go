package accumulation

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

type OuterAccumulationInput struct {
	GasLimit                     types.Gas                 // g    gas-limit
	DeferredTransfers            []types.DeferredTransfer  // t    deferred transfers
	WorkReports                  []types.WorkReport        // r    a sequence of work-reports
	InitPartialStateSet          types.PartialStateSet     // e    an initial partial-state
	ServicesWithFreeAccumulation types.AlwaysAccumulateMap // f    a dictionary of services enjoying free accumulation
}

type OuterAccumulationOutput struct {
	NumberOfWorkResultsAccumulated types.U64
	PartialStateSet                types.PartialStateSet          // e a posterior state-context
	AccumulatedServiceOutput       types.AccumulatedServiceOutput // b service/hash pairs
	ServiceGasUsedList             types.ServiceGasUsedList       // u service/gas pairs
}

type ParallelizedAccumulationInput struct {
	PartialStateSet     types.PartialStateSet     // e   initial state-context
	DeferredTransfers   []types.DeferredTransfer  // t   deferred transfers
	WorkReports         []types.WorkReport        // r   a sequence of work-reports
	AlwaysAccumulateMap types.AlwaysAccumulateMap // f   a dictionary of privileged always-accumulate services
}

type ParallelizedAccumulationOutput struct {
	PartialStateSet          types.PartialStateSet          // e a posterior state-context
	DeferredTransfers        []types.DeferredTransfer       // t resultant deferred transfers
	AccumulatedServiceOutput types.AccumulatedServiceOutput // b service/hash pairs
	ServiceGasUsedList       types.ServiceGasUsedList       // u service/gas pairs
}

type SingleServiceAccumulationInput struct {
	PartialStateSet     types.PartialStateSet     // e   initial state-context
	DeferredTransfers   []types.DeferredTransfer  // t   deferred transfers
	WorkReports         []types.WorkReport        // r   a sequence of work-reports
	AlwaysAccumulateMap types.AlwaysAccumulateMap // f   a dictionary of privileged always-accumulate services
	ServiceId           types.ServiceId           // s   a service index
	UnmatchedKeyVals    types.StateKeyVals        // storage key-value pairs that were not matched yet
}

type SingleServiceAccumulationOutput struct {
	PartialStateSet    types.PartialStateSet    // an alterations state-context
	DeferredTransfers  []types.DeferredTransfer // a sequence of transfers
	AccumulationOutput *types.OpaqueHash        // a possible accumulation-output
	GasUsed            types.Gas                // the actual PVM gas used
	ServiceBlobs       types.ServiceBlobs       // a hash service pair of the accumulated service
	UnmatchedKeyVals   types.StateKeyVals       // storage key-value pairs that were not matched yet
}
