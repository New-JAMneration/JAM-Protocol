package extrinsic

import (
	"sort"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// GuaranteeCredentialsController is a struct that contains a slice of GuaranteeCredentials (for controller logic)
type GuaranteeCredentialsController struct {
	Credentials []types.ValidatorSignature
}

// NewGuaranteeCredentialsController creates a new GuaranteeCredentialsController (Constructor)
func NewGuaranteeCredentialsController() *GuaranteeCredentialsController {
	return &GuaranteeCredentialsController{
		Credentials: make([]types.ValidatorSignature, 0),
	}
}

// Set sets the GuaranteeCredentials slice
func (c *GuaranteeCredentialsController) Set(cToSet []types.ValidatorSignature) {
	c.Credentials = cToSet
}

// Len returns the length of the slice
func (c *GuaranteeCredentialsController) Len() int {
	return len(c.Credentials)
}

// Less returns true if the index i is less than the index j
func (c *GuaranteeCredentialsController) Less(i, j int) bool {
	return c.Credentials[i].ValidatorIndex < c.Credentials[j].ValidatorIndex
}

// Swap swaps the index i with the index j
func (c *GuaranteeCredentialsController) Swap(i, j int) {
	c.Credentials[i], c.Credentials[j] = c.Credentials[j], c.Credentials[i]
}

// Sort sorts the slice
func (c *GuaranteeCredentialsController) Sort() {
	sort.Slice(c.Credentials, func(i, j int) bool {
		return c.Less(i, j)
	})
}

// Add adds a new GuaranteeCredentials to the GuaranteeCredentials slice.
// It also removes the duplicates from the slice.
func (c *GuaranteeCredentialsController) Add(newReportGuaranteeCredentials types.ValidatorSignature) []types.ValidatorSignature {
	c.Credentials = append(c.Credentials, newReportGuaranteeCredentials)
	return c.RemoveDuplicates()
}

// RemoveDuplicates removes the duplicates from the ValidatorSignature slice.
// It uses the Sort function to sort the slice first.
// And uses double pointers to remove the duplicates.
func (c *GuaranteeCredentialsController) RemoveDuplicates() []types.ValidatorSignature {
	if len(c.Credentials) == 0 {
		return c.Credentials
	}

	// First, sort the slice
	c.Sort()

	// Then, remove the duplicates
	j := 0

	for i := 1; i < len(c.Credentials); i++ {
		if c.Credentials[i].ValidatorIndex != c.Credentials[j].ValidatorIndex {
			j++
			c.Credentials[j] = c.Credentials[i]
		}
	}

	// Remove the duplicates
	return c.Credentials[:j+1]
}
