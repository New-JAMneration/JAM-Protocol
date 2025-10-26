package memory

import (
	"sort"
	"strings"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
)

type iterator struct {
	index     int
	currKey   string
	currValue []byte
	keys      []string
	values    [][]byte
}

// NewIterator creates a new iterator for the given prefix.
// The start key is inclusive.
func (db *memoryDB) NewIterator(prefix []byte, start []byte) (database.Iterator, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	prefixString := string(prefix)
	startString := string(append(prefix, start...))

	var keys []string

	// Collect all keys in the range [start, end)
	for key := range db.data {
		if !strings.HasPrefix(key, prefixString) {
			continue
		}
		if strings.Compare(key, startString) >= 0 {
			keys = append(keys, key)
		}
	}

	sort.Strings(keys)

	// Collect values by sorted keys
	values := make([][]byte, len(keys))
	for i, key := range keys {
		// Make a copy of the value
		v := db.data[key]
		valueCopy := make([]byte, len(v))
		copy(valueCopy, v)
		values[i] = valueCopy
	}

	return &iterator{
		index:  -1, // Start at -1 so the first Next() call moves to index 0
		keys:   keys,
		values: values,
	}, nil
}

// Next advances the iterator to the next key/value pair.
func (iter *iterator) Next() bool {
	iter.index++
	if iter.index >= len(iter.keys) {
		return false
	}
	iter.currKey = iter.keys[iter.index]
	iter.currValue = iter.values[iter.index]
	return true
}

// Key returns the current key.
// The returned slice is only valid until the next call to `Next()`, and should not be modified.
func (iter *iterator) Key() []byte {
	if iter.index < 0 || iter.index >= len(iter.keys) {
		return nil
	}
	return []byte(iter.currKey)
}

// Value returns the current value.
// The returned slice is only valid until the next call to `Next()`, and should not be modified.
func (iter *iterator) Value() []byte {
	if iter.index < 0 || iter.index >= len(iter.values) {
		return nil
	}
	return iter.currValue
}

// Error returns any error encountered during iteration.
// In-memory iterator does not produce errors, so this always returns nil.
func (iter *iterator) Error() error {
	return nil
}

// Close closes the iterator.
// It is safe to call Close multiple times.
func (iter *iterator) Close() error {
	iter.index = -1
	iter.keys = nil
	iter.values = nil
	return nil
}
