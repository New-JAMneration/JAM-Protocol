package memory

import (
	"bytes"
	"sync"
)

type KVStore struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func New() *KVStore {
	return &KVStore{
		data: make(map[string][]byte),
	}
}

func (kv *KVStore) Has(key []byte) (bool, error) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()

	_, exists := kv.data[string(key)]
	return exists, nil
}

func (kv *KVStore) Get(key []byte) ([]byte, error) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()

	value, exists := kv.data[string(key)]
	if !exists {
		return nil, nil
	}

	result := make([]byte, len(value))
	copy(result, value)
	return result, nil
}

func (kv *KVStore) Set(key, value []byte) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)
	kv.data[string(key)] = valueCopy
	return nil
}

func (kv *KVStore) Delete(key []byte) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	delete(kv.data, string(key))
	return nil
}

func (kv *KVStore) DeleteRange(start, end []byte) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	for k := range kv.data {
		keyBytes := []byte(k)
		if bytes.Compare(keyBytes, start) >= 0 && bytes.Compare(keyBytes, end) < 0 {
			delete(kv.data, k)
		}
	}
	return nil
}

func (kv *KVStore) Close() error {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	kv.data = nil
	return nil
}
