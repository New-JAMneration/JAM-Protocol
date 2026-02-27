package auditing

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// BundleFetcher abstracts how an auditor obtains the original WorkPackageBundle
// for a given WorkReport. The node networking layer should provide a concrete
// implementation that:
//
//  1. (Fast path) Requests the full bundle from a guarantor via CE protocol,
//     identified by the guarantor validator indices in the ReportGuarantee.
//  2. (Slow path) Reconstructs the bundle from ≥342 erasure-coded chunks
//     fetched from validators, using the erasure-root from WorkPackageSpec.
//
// Both paths must verify the received bundle against the erasure-root
// before returning. See GP §17.16.
type BundleFetcher interface {
	// FetchBundle retrieves the encoded WorkPackageBundle bytes for the given
	// work report. The erasureRoot from report.PackageSpec.ErasureRoot should
	// be used to verify the fetched data.
	//
	// Returns the raw encoded bundle bytes suitable for decoding into
	// types.WorkPackageBundle, or an error if the bundle cannot be obtained.
	FetchBundle(report types.WorkReport) ([]byte, error)
}

// StubBundleFetcher is a placeholder implementation that always returns an error.
// Replace this with a real implementation once the node networking layer (CE
// protocol for bundle requests) is integrated from feat/jam-np-ce-handler.
//
// Node integration TODO:
//   - Identify guarantor peers from ReportGuarantee.Signatures[].ValidatorIndex
//   - Send bundle request CE with report.PackageSpec.ErasureRoot
//   - Receive and verify bundle against PackageSpec.ErasureRoot
//     (use work_package.ComputeErasureRoot to recompute and compare)
//   - Fallback to erasure-coded chunk reconstruction if guarantor is unreachable
type StubBundleFetcher struct{}

func (f *StubBundleFetcher) FetchBundle(report types.WorkReport) ([]byte, error) {
	return nil, fmt.Errorf(
		"bundle fetching not yet implemented: requires node networking layer (CE protocol) "+
			"to request bundle for work-package %x from guarantors",
		report.PackageSpec.Hash,
	)
}
