package pebble

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	"github.com/cockroachdb/pebble"
)

type iterator struct {
	inner  *pebble.Iterator
	isHead bool
	closed bool
}

// NewIterator creates a new iterator for the given prefix.
// The start key is inclusive, and the end key is exclusive.
func (db *pebbleDB) NewIterator(prefix []byte, start []byte) (database.Iterator, error) {
	// prefix iterator upper bound calculation
	// https://github.com/cockroachdb/pebble/blob/ffc306f908df470254d953bf865aca1c94e49271/iterator_example_test.go#L44
	keyUpperBound := func(b []byte) []byte {
		end := make([]byte, len(b))
		copy(end, b)
		for i := len(end) - 1; i >= 0; i-- {
			end[i] = end[i] + 1
			if end[i] != 0 {
				return end[:i+1]
			}
		}
		return nil // no upper-bound
	}

	iter, err := db.inner.NewIter(&pebble.IterOptions{
		LowerBound: append(prefix, start...),
		UpperBound: keyUpperBound(prefix),
	})
	if err != nil {
		return nil, err
	}

	iter.First()
	return &iterator{
		inner:  iter,
		isHead: true,
		closed: false,
	}, nil
}

// Next advances the iterator to the next key/value pair.
func (it *iterator) Next() bool {
	if it.isHead {
		it.isHead = false
		return it.inner.Valid()
	}
	return it.inner.Next()
}

// Key returns the current key.
// The returned slice is only valid until the next call to `Next()`, and should not be modified.
func (iter *iterator) Key() []byte {
	return iter.inner.Key()
}

// Value returns the current value.
// The returned slice is only valid until the next call to `Next()`, and should not be modified.
func (it *iterator) Value() []byte {
	return it.inner.Value()
}

// Error returns the error encountered during iteration, if any.
func (it *iterator) Error() error {
	return it.inner.Error()
}

// Close closes the iterator.
// It is safe to call Close multiple times.
func (it *iterator) Close() error {
	if it.closed {
		return nil
	}
	it.closed = true
	return it.inner.Close()
}
