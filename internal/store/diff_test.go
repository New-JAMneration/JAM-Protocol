package store_test

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/stretchr/testify/assert"
)

func kv(keyByte byte, value ...byte) types.StateKeyVal {
	var key types.StateKey
	key[0] = keyByte
	return types.StateKeyVal{Key: key, Value: value}
}

func TestDiffSortedKeyVals_NoDiff(t *testing.T) {
	kvs := types.StateKeyVals{kv(1, 0xAA), kv(2, 0xBB)}
	result := store.DiffSortedKeyVals(kvs, kvs)
	assert.Empty(t, result)
}

func TestDiffSortedKeyVals_AllInserted(t *testing.T) {
	prior := types.StateKeyVals{}
	current := types.StateKeyVals{kv(1, 0xAA), kv(2, 0xBB)}
	result := store.DiffSortedKeyVals(prior, current)

	assert.Len(t, result, 2)
	assert.Equal(t, types.StateKey{1}, result[0].Key)
	assert.Equal(t, []byte{0xAA}, result[0].NewValue)
	assert.False(t, result[0].IsDelete)
	assert.Equal(t, types.StateKey{2}, result[1].Key)
	assert.False(t, result[1].IsDelete)
}

func TestDiffSortedKeyVals_AllDeleted(t *testing.T) {
	prior := types.StateKeyVals{kv(1, 0xAA), kv(2, 0xBB)}
	current := types.StateKeyVals{}
	result := store.DiffSortedKeyVals(prior, current)

	assert.Len(t, result, 2)
	assert.Equal(t, types.StateKey{1}, result[0].Key)
	assert.True(t, result[0].IsDelete)
	assert.Nil(t, result[0].NewValue)
	assert.Equal(t, types.StateKey{2}, result[1].Key)
	assert.True(t, result[1].IsDelete)
}

func TestDiffSortedKeyVals_Modified(t *testing.T) {
	prior := types.StateKeyVals{kv(1, 0xAA), kv(2, 0xBB)}
	current := types.StateKeyVals{kv(1, 0xAA), kv(2, 0xCC)}
	result := store.DiffSortedKeyVals(prior, current)

	assert.Len(t, result, 1)
	assert.Equal(t, types.StateKey{2}, result[0].Key)
	assert.Equal(t, []byte{0xCC}, result[0].NewValue)
	assert.False(t, result[0].IsDelete)
}

func TestDiffSortedKeyVals_Mixed(t *testing.T) {
	prior := types.StateKeyVals{
		kv(1, 0xAA),
		kv(2, 0xBB),
		kv(4, 0xDD),
	}
	current := types.StateKeyVals{
		kv(1, 0xAA), // unchanged
		kv(2, 0xFF), // modified
		kv(3, 0xCC), // inserted
		// key=4 deleted
	}
	result := store.DiffSortedKeyVals(prior, current)

	assert.Len(t, result, 3)

	// modified: key=2
	assert.Equal(t, types.StateKey{2}, result[0].Key)
	assert.Equal(t, []byte{0xFF}, result[0].NewValue)
	assert.False(t, result[0].IsDelete)

	// inserted: key=3
	assert.Equal(t, types.StateKey{3}, result[1].Key)
	assert.Equal(t, []byte{0xCC}, result[1].NewValue)
	assert.False(t, result[1].IsDelete)

	// deleted: key=4
	assert.Equal(t, types.StateKey{4}, result[2].Key)
	assert.True(t, result[2].IsDelete)
}

func TestDiffSortedKeyVals_BothEmpty(t *testing.T) {
	result := store.DiffSortedKeyVals(nil, nil)
	assert.Empty(t, result)
}

func TestDiffSortedKeyVals_LargeValueChange(t *testing.T) {
	largeOld := make([]byte, 64)
	largeNew := make([]byte, 64)
	largeNew[63] = 1

	prior := types.StateKeyVals{{Key: types.StateKey{0x10}, Value: largeOld}}
	current := types.StateKeyVals{{Key: types.StateKey{0x10}, Value: largeNew}}
	result := store.DiffSortedKeyVals(prior, current)

	assert.Len(t, result, 1)
	assert.Equal(t, largeNew, result[0].NewValue)
	assert.False(t, result[0].IsDelete)
}
