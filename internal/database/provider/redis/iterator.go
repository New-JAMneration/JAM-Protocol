package redis

import (
	"strings"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	"github.com/go-redis/redis"
)

type iterator struct {
	index     int
	currKey   []byte
	currValue []byte
	keys      [][]byte
	values    [][]byte
}

// NewIterator creates a new iterator for the given prefix. The start key is inclusive.
func (db *redisDB) NewIterator(prefix []byte, start []byte) (database.Iterator, error) {
	buf := make([]byte, 0, len(prefix)+len(start))
	buf = append(buf, prefix...)
	buf = append(buf, start...)
	startString := string(buf)

	var nextCursor uint64
	var allKeys []string
	var err error

	pattern := startString + "*"

	for {
		var keys []string
		keys, nextCursor, err = db.client.Scan(nextCursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}

		// Filter keys that match the prefix
		for _, key := range keys {
			if strings.HasPrefix(key, startString) {
				allKeys = append(allKeys, key)
			}
		}

		if nextCursor == 0 {
			break
		}
	}

	var keys [][]byte
	var values [][]byte
	for _, key := range allKeys {
		value, err := db.client.Get(key).Bytes()
		if err != nil && err != redis.Nil {
			return nil, err
		}
		if err == redis.Nil {
			// Key was deleted between scan and get, skip it
			continue
		}
		keys = append(keys, []byte(key))
		values = append(values, value)
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
	return iter.currKey
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
// Redis iterator does not produce errors, so this always returns nil.
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
