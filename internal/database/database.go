package database

import (
	"io"
)

// Database is all interfaces for underlying key-value database.
type Database interface {
	Reader
	Writer
	Batcher
	Iterable
	io.Closer
}

// Reader defines read-only operations for a key-value database.
type Reader interface {
	Has(key []byte) (bool, error)
	Get(key []byte) ([]byte, bool, error)
}

// Writer defines write-only operations for a key-value database.
type Writer interface {
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
	Writer

	// Commit writes all changes buffered in the batch to the underlying database.
	Commit() error
}

// Iterable defines new iterator creation for a key-value database.
type Iterable interface {
	NewIterator(prefix []byte, start []byte) (Iterator, error)
}

// Iterator defines the interface for iterating over key-value pairs in the database.
type Iterator interface {
	Next() bool
	Key() []byte
	Value() []byte
	Error() error
	Close() error
}
