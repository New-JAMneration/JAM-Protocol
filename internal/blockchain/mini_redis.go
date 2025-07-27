package blockchain

import "github.com/New-JAMneration/JAM-Protocol/internal/store"

// CloseMiniRedis is a test helper used by some networking tests.
// Some tests toggle USE_MINI_REDIS and expect cleanup via this hook.
func CloseMiniRedis() { store.CloseMiniRedis() }
