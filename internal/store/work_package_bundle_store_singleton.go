package store

import "sync"

var (
	wpBundleStoreMu sync.RWMutex
	wpBundleStore   *WorkPackageBundleStore
)

func GetWorkPackageBundleStore() *WorkPackageBundleStore {
	wpBundleStoreMu.RLock()
	defer wpBundleStoreMu.RUnlock()
	return wpBundleStore
}

func SetWorkPackageBundleStore(s *WorkPackageBundleStore) {
	wpBundleStoreMu.Lock()
	defer wpBundleStoreMu.Unlock()
	wpBundleStore = s
}

func ResetWorkPackageBundleStore() {
	SetWorkPackageBundleStore(nil)
}
