package pebble

import (
	"runtime"

	"github.com/cockroachdb/pebble"
)

type KVStore struct {
	datadir   string
	inner     *pebble.DB
	writeOpts *pebble.WriteOptions
}

func New(datadir string, readOnly bool) (*KVStore, error) {
	// TODO: Set other options for optimizations, enable flexible config if needed.
	opt := &pebble.Options{
		// Default compaction concurrency is single-threaded, use all available logical cores
		// for speeding up compactions.
		MaxConcurrentCompactions: runtime.NumCPU,
		ReadOnly:                 readOnly,
	}
	pebbleDB, err := pebble.Open(datadir, opt)
	if err != nil {
		return nil, err
	}
	engine := &KVStore{
		datadir:   datadir,
		inner:     pebbleDB,
		writeOpts: pebble.NoSync,
	}
	return engine, nil
}

func (kv *KVStore) Has(key []byte) (bool, error) {
	_, closer, err := kv.inner.Get(key)
	if err == pebble.ErrNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}
	if err = closer.Close(); err != nil {
		return false, err
	}
	return true, nil
}

func (kv *KVStore) Get(key []byte) ([]byte, error) {
	value, closer, err := kv.inner.Get(key)
	if err != nil {
		return nil, err
	}
	if err = closer.Close(); err != nil {
		return nil, err
	}
	return value, nil
}

func (kv *KVStore) Set(key, value []byte) error {
	return kv.inner.Set(key, value, kv.writeOpts)
}

func (kv *KVStore) Delete(key []byte) error {
	return kv.inner.Delete(key, kv.writeOpts)
}

func (kv *KVStore) DeleteRange(start, end []byte) error {
	return kv.inner.DeleteRange(start, end, kv.writeOpts)
}

func (kv *KVStore) Close() error {
	return kv.inner.Close()
}
