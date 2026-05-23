//go:build !dev_assert

// No-op stub for AssertConsistency in regular builds. See
// assert_consistency.go for the real implementation enabled via the
// `dev_assert` build tag.
package service_account

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

// AssertConsistency is a no-op in regular builds.
func AssertConsistency(_ types.ServiceID, _ types.ServiceAccount) error {
	return nil
}
