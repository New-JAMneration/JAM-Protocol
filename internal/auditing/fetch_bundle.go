package auditing

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// BundleFetcher abstracts how an auditor obtains the original WorkPackageBundle.
// Node layer should implement this with:
//   - Fast path: request bundle from guarantor via CE
//   - Slow path: reconstruct from ≥342 erasure-coded chunks
//
// Both paths must verify against PackageSpec.ErasureRoot. See GP §17.16.
type BundleFetcher interface {
	FetchBundle(report types.WorkReport) ([]byte, error)
}

// NODE-TODO [bundle fetch]: Replace with real impl from feat/jam-np-ce-handler.
// Steps: find guarantor peers (ReportGuarantee.Signatures[].ValidatorIndex) →
// request bundle via CE → verify with work_package.ComputeErasureRoot →
// fallback to erasure reconstruction if unreachable.
type StubBundleFetcher struct{}

func (f *StubBundleFetcher) FetchBundle(report types.WorkReport) ([]byte, error) {
	return nil, fmt.Errorf(
		"bundle fetching not yet implemented: requires node networking layer (CE protocol) "+
			"to request bundle for work-package %x from guarantors",
		report.PackageSpec.Hash,
	)
}
