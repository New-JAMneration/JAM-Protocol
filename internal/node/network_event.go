package node

type NetworkEvent int

const (
	PeerAdded NetworkEvent = iota
	PeerUpdated
	// BulkSyncCompleted
)
