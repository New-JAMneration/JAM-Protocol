package memory

import (
	"bytes"
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
)

type memoryDB struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewDatabase() database.Database {
	return &memoryDB{
		mu:   sync.RWMutex{},
		data: make(map[string][]byte),
	}
}

func (db *memoryDB) Has(key []byte) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	_, exists := db.data[string(key)]
	return exists, nil
}

func (db *memoryDB) Get(key []byte) ([]byte, bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	value, exists := db.data[string(key)]
	if !exists {
		return nil, false, nil
	}

	result := make([]byte, len(value))
	copy(result, value)
	return result, true, nil
}

func (db *memoryDB) Put(key, value []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)
	db.data[string(key)] = valueCopy
	return nil
}

func (db *memoryDB) Delete(key []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	delete(db.data, string(key))
	return nil
}

func (db *memoryDB) DeleteRange(start, end []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for k := range db.data {
		keyBytes := []byte(k)
		if bytes.Compare(keyBytes, start) >= 0 && bytes.Compare(keyBytes, end) < 0 {
			delete(db.data, k)
		}
	}
	return nil
}

func (db *memoryDB) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.data = nil
	return nil
}
