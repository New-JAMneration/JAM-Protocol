package eventbus

type EventType string

const (
	EventNewBlock                 EventType = "newBlock"
	EventFinalizedBlock           EventType = "finalizedBlock"
	EventImportedBlock            EventType = "importedBlock"
	EventSyncStateChanged         EventType = "syncStateChanged"
	EventStatisticsChanged        EventType = "statisticsChanged"
	EventServiceDataChanged       EventType = "serviceDataChanged"
	EventServiceValueChanged      EventType = "serviceValueChanged"
	EventServicePreimageChanged   EventType = "servicePreimageChanged"
	EventServiceRequestChanged    EventType = "serviceRequestChanged"
	EventWorkPackageStatusChanged EventType = "workPackageStatusChanged"
)

type Event struct {
	Type EventType
	Data interface{}
}

type BlockEvent struct {
	HeaderHash string `json:"header_hash"`
	Slot       uint64 `json:"slot"`
}

type SyncStateEvent struct {
	NumPeers int    `json:"num_peers"`
	Status   string `json:"status"`
}

type ChainSubscriptionUpdate struct {
	HeaderHash string      `json:"header_hash"`
	Slot       uint64      `json:"slot"`
	Value      interface{} `json:"value"`
}

type ServiceEvent struct {
	ServiceID  string      `json:"service_id"`
	Key        *string     `json:"key,omitempty"`
	Hash       *string     `json:"hash,omitempty"`
	Length     *uint32     `json:"length,omitempty"`
	HeaderHash string      `json:"header_hash"`
	Slot       uint64      `json:"slot"`
	Value      interface{} `json:"value"`
}

type WorkPackageStatusEvent struct {
	Hash       string      `json:"hash"`
	Anchor     string      `json:"anchor"`
	HeaderHash string      `json:"header_hash"`
	Slot       uint64      `json:"slot"`
	Status     interface{} `json:"status"`
}
