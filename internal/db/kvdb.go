package db

import "io"

type KeyValueDB interface {
	KeyValueReader
	KeyValueWriter
	io.Closer
}

type KeyValueReader interface {
	Has(key []byte) (bool, error)
	Get(key []byte) ([]byte, error)
}

type KeyValueWriter interface {
	Set(key, value []byte) error
	Delete(key []byte) error
	DeleteRange(start, end []byte) error
}
