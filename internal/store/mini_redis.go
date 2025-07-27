package store

// CloseMiniRedis closes the embedded in-memory redis (when USE_MINI_REDIS=true).
// It is safe to call even if the backend was never initialized.
func CloseMiniRedis() {
	resetRedisBackend()
}
