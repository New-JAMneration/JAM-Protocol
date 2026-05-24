//go:build dev_assert

// Dev-only consistency assertion for the global-KV transition.
//
// The legacy StorageDict / LookupDict fields have been removed; this
// assertion is now a no-op stub matching assert_consistency_stub.go.
// Retained to avoid breaking the build under `-tags dev_assert`.
package service_account

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

// AssertConsistency is a no-op now that the legacy maps have been removed.
func AssertConsistency(_ types.ServiceID, _ types.ServiceAccount) error {
	return nil
}
