package database

import "io"

// Database is all interfaces for underlying key-value database.
type Database interface {
	KeyValueReader
	KeyValueWriter
	Batcher
	io.Closer
}

// KeyValueReader defines read-only operations for a key-value database.
type KeyValueReader interface {
	Has(key []byte) (bool, error)
	Get(key []byte) ([]byte, bool, error)
}

// KeyValueWriter defines write-only operations for a key-value database.
type KeyValueWriter interface {
	Put(key, value []byte) error
	Delete(key []byte) error
	DeleteRange(start, end []byte) error
}

// Batcher defines batch write operations for a key-value database.
type Batcher interface {
	NewBatch() Batch
}

// Batch is a set of write-only operations.
// The operations are buffered until Commit is explicitely called.
type Batch interface {
	KeyValueWriter

	// Commit writes all changes buffered in the batch to the underlying database.
	Commit() error
}
