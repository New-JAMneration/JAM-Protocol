package merklization

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// LeafHashCache is a get-or-compute callback for leaf hashes.
// It returns the leaf hash for (key, value);
// On cache miss the implementation: computes it, stores it, and returns it.
type LeafHashCache func(key types.StateKey, value []byte) (leafHash types.OpaqueHash)
