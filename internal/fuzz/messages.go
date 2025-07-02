package fuzz

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type MessageType uint8

const (
	MessageType_PeerInfo    MessageType = 0
	MessageType_ImportBlock MessageType = 1
	MessageType_SetState    MessageType = 2
	MessageType_GetState    MessageType = 3
	MessageType_State       MessageType = 4
	MessageType_StateRoot   MessageType = 5
)

type (
	TrieKey [31]byte

	Version struct {
		Major uint8
		Minor uint8
		Patch uint8
	}

	PeerInfo struct {
		Name       string
		AppVersion Version
		JamVersion Version
	}

	KeyValue struct {
		Key   TrieKey
		Value []byte
	}

	State []KeyValue

	ImportBlock types.Block

	SetState struct {
		Header types.Header
		State  State
	}

	GetState types.HeaderHash

	StateRoot types.StateRoot

	Message struct {
		Type MessageType

		PeerInfo    *PeerInfo
		ImportBlock *ImportBlock
		SetState    *SetState
		GetState    *GetState
		State       *State
		StateRoot   *StateRoot
	}
)
