package accumulation

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

// (12.9)
// the mapping function P which extracts the corresponding work-package hashes from a set of work-reports
// FIXME: naming issue
func MappingFunction(workReports []types.WorkReport) []types.WorkPackageHash {
	workPackageHashes := make([]types.WorkPackageHash, 0)
	for _, workReport := range workReports {
		workPackageHashes = append(workPackageHashes, workReport.PackageSpec.Hash)
	}

	return workPackageHashes
}

// (12.7)
// the queue-editing function E,
// which is essentially a mutator functino for items such as those of ϑ,
// parameterized by sets of now-accumulated work-package hashes (those in ξ)
// It is used to update queues  of work-reports when some of them are accumulated.
// Functionally, it removes all entries whose work-report's hash is in the set provided as a parameter,
// and removes any dependencies which apper in said set.
// FIXME: argument type: types.AccumulatedHistory or []types.WorkPackageHash
// naming issue
func QueueEditingFunction(readyQueueItem types.ReadyQueueItem, workPackageHashes []types.WorkPackageHash) types.ReadyQueueItem {
	return types.ReadyQueueItem{}
}

// (12.16)
// outer accumulation function
func OuterAccumulation(input OuterAccumulationInput) (OuterAccumulationOutput, error) {
	var err error

	return OuterAccumulationOutput{}, err
}

// (12.17)
// parallelized accumulation function
func ParallelizedAccumulation(input ParallelizedAccumulationInput) (ParallelizedAccumulationOutput, error) {
	var err error

	return ParallelizedAccumulationOutput{}, err
}

// (12.19)
// single-service accumulation function
func SingleServiceAccumulation(input SingleServiceAccumulationInput) (SingleServiceAccumulationOutput, error) {
	var err error

	return SingleServiceAccumulationOutput{}, err
}
