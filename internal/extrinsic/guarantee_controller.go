package extrinsic

import (
	"sort"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// GuaranteeController is a struct that contains a slice of ReportGuarantee (for controller logic)
type GuaranteeController struct {
	Guarantees []types.ReportGuarantee
}

// NewGuaranteeController creates a new GuaranteeController (Constructor)
func NewGuaranteeController() *GuaranteeController {
	return &GuaranteeController{
		Guarantees: make([]types.ReportGuarantee, 0),
	}
}

// Set sets the ReportGuarantee slice
func (g *GuaranteeController) Set(gToSet []types.ReportGuarantee) {
	g.Guarantees = gToSet
}

// Len returns the length of the slice
func (r *GuaranteeController) Len() int {
	return len(r.Guarantees)
}

// Less returns true if the index i is less than the index j
func (r *GuaranteeController) Less(i, j int) bool {
	return r.Guarantees[i].Report.CoreIndex < r.Guarantees[j].Report.CoreIndex
}

// Swap swaps the index i with the index j
func (r *GuaranteeController) Swap(i, j int) {
	r.Guarantees[i], r.Guarantees[j] = r.Guarantees[j], r.Guarantees[i]
}

// Sort sorts the slice
func (r *GuaranteeController) Sort() {
	sort.Slice(r.Guarantees, func(i, j int) bool {
		return r.Less(i, j)
	})
}

// Add adds a new Guarantee to the ReportGuarantee slice.
func (r *GuaranteeController) Add(newReportGuarantee types.ReportGuarantee) {
	r.Guarantees = append(r.Guarantees, newReportGuarantee)
	r.Sort()
}
