package keystore

import "errors"

var (
	ErrInvalidKeyType = errors.New("invalid key type")
	ErrKeyNotFound    = errors.New("key not found")
	ErrKeyExists      = errors.New("key already exists")
)

// Storage defines the interface for keystore storage backends.
// This interface is designed to be database-agnostic, allowing
// implementations for Redis, Pebble, Memory, or any other database.
type Storage interface {
	// Basic key-value operations
	Put(key string, value []byte) error
	Get(key string) ([]byte, error) // Returns ErrKeyNotFound if key doesn't exist
	Delete(key string) error

	// Set operations (for maintaining public key lists)
	SetAdd(key string, members ...[]byte) error
	SetRemove(key string, members ...[]byte) error
	SetMembers(key string) ([][]byte, error)
	SetIsMember(key string, member []byte) (bool, error)

	GetMultiple(keys []string) (map[string][]byte, error)

	Begin() (Transaction, error)
}

// Transaction represents a storage transaction
type Transaction interface {
	Put(key string, value []byte) error
	Delete(key string) error
	SetAdd(key string, members ...[]byte) error
	SetRemove(key string, members ...[]byte) error
	Commit() error
	Rollback() error
}
