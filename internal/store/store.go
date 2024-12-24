package store

import "sync"

var (
    // initOnce ensures the store is initialized only once
    initOnce sync.Once
    // globalStore holds the singleton instance of Store
    globalStore *Store
)

// Store represents a thread-safe global state container
type Store struct {
    mu sync.RWMutex
    // data is a map that can store any type of value
    data map[string]interface{}
}

// GetInstance returns the singleton instance of Store.
// If the instance doesn't exist, it creates one.
func GetInstance() *Store {
    initOnce.Do(func() {
        globalStore = &Store{
            data: make(map[string]interface{}),
        }
    })
    return globalStore
}

// Set stores a value with the given key in the store.
// This operation is thread-safe.
func (s *Store) Set(key string, value interface{}) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.data[key] = value
}

// Get retrieves a value by its key from the store.
// Returns the value and a boolean indicating if the key exists.
// This operation is thread-safe.
func (s *Store) Get(key string) (interface{}, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    value, exists := s.data[key]
    return value, exists
}
