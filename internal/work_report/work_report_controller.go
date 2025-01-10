package work_report

import (
	"fmt"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type WorkReportController struct {
	WorkReport *types.WorkReport
	/*
		type WorkReport struct {
			PackageSpec       WorkPackageSpec   `json:"package_spec"`
			Context           RefineContext     `json:"context"`
			CoreIndex         CoreIndex         `json:"core_index,omitempty"`
			AuthorizerHash    OpaqueHash        `json:"authorizer_hash,omitempty"`
			AuthOutput        ByteSequence      `json:"auth_output,omitempty"`
			SegmentRootLookup SegmentRootLookup `json:"segment_root_lookup,omitempty"`
			Results           []WorkResult      `json:"results,omitempty"`
		}
	*/
}

// NewWorkReportController creates a new WorkReportController
func NewWorkReportController(workReport *types.WorkReport) *WorkReportController {
	return &WorkReportController{
		WorkReport: workReport,
	}
}

// ValidateWorkItemsNumbers checks if the number of work items is between 1 and 4 | Eq. 11.2
func (w *WorkReportController) ValidateWorkItemsNumbers() error {
	if len(w.WorkReport.Results) < 1 || len(w.WorkReport.Results) > types.I {
		return fmt.Errorf("WorkReport Results must have between 1 and 4 items, but got %d", len(w.WorkReport.Results))
	}
	return nil
}

// ValidateLookupDictAndPrerequisites checks the number of SegmentRootLookup and Prerequisites < J | Eq. 11.3
func (w *WorkReportController) ValidateLookupDictAndPrerequisites() error {
	if len(w.WorkReport.SegmentRootLookup)+len(w.WorkReport.Context.Prerequisites) > types.J {
		return fmt.Errorf("SegmentRootLookup and Prerequisites must have a total at most %d, but got %d", types.J, len(w.WorkReport.SegmentRootLookup)+len(w.WorkReport.Context.Prerequisites))
	}
	return nil
}

// ValidateOutputSize checks the total size of the output | Eq. 11.8
func (w *WorkReportController) ValidateOutputSize() error {
	totalSize := len(w.WorkReport.AuthOutput)
	for _, result := range w.WorkReport.Results {
		for _, outputs := range result.Result {
			totalSize += len(outputs)
		}
	}

	if totalSize > types.WorkReportOutputBlobsMaximumSize {
		return fmt.Errorf("total size exceeds %d bytes", types.WorkReportOutputBlobsMaximumSize)
	}
	return nil
}
