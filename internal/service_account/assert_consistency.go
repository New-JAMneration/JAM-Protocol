//go:build dev_assert

// Dev-only consistency assertion for the global-KV transition.
//
// During Step 1-7 of the refactor the legacy StorageDict / LookupDict and
// the new globalKV / counters coexist. This file provides AssertConsistency
// which can be called from PVM host calls / accumulation hooks to verify
// that the two representations agree.
//
// Compiled in only with `-tags dev_assert`; in regular builds the function
// is replaced by a no-op stub (see assert_consistency_stub.go).
package service_account

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// AssertConsistency panics if the counters disagree with the legacy maps.
// Returns nil on success and an error describing the discrepancy otherwise.
func AssertConsistency(serviceID types.ServiceID, account types.ServiceAccount) error {
	expectedItems := uint32(2*len(account.LookupDict) + len(account.StorageDict))
	if got := account.GetTotalNumberOfItems(); got != expectedItems {
		return fmt.Errorf("service %d: totalNumberOfItems mismatch: counter=%d legacy=%d", serviceID, got, expectedItems)
	}

	var expectedOctets uint64
	for key := range account.LookupDict {
		expectedOctets += 81 + uint64(key.Length)
	}
	for k, v := range account.StorageDict {
		expectedOctets += 34 + uint64(len(k)) + uint64(len(v))
	}
	if got := account.GetTotalNumberOfOctets(); got != expectedOctets {
		return fmt.Errorf("service %d: totalNumberOfOctets mismatch: counter=%d legacy=%d", serviceID, got, expectedOctets)
	}

	return nil
}
