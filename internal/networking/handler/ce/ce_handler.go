package ce

import (
	"errors"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
)

// CEHandlerFunc defines the function signature for handling a specific CE request.
// The provided stream's payload (after the protocol ID) is available for further processing.
type CEHandlerFunc func(blockchain blockchain.Blockchain, stream *quic.Stream) error

// CEHandler is a generic handler for Common Ephemeral (CE) streams.
// It dispatches requests based on the first byte (protocol ID).
type CEHandler struct {
	handlers map[uint8]CEHandlerFunc
}

// New creates a new CEHandler instance.
func New() *CEHandler {
	return &CEHandler{
		handlers: make(map[uint8]CEHandlerFunc),
	}
}

// Register associates a protocol ID with a handler function.
func (h *CEHandler) Register(protoID uint8, handler CEHandlerFunc) {
	h.handlers[protoID] = handler
}

// HandleStream dispatches to the registered CE handler using a stream kind that
// has already been consumed by the upstream stream dispatcher.
func (h *CEHandler) HandleStream(blockchain blockchain.Blockchain, protoID uint8, stream *quic.Stream) error {
	handler, ok := h.handlers[protoID]
	if !ok {
		return errors.New("unsupported CE request protocol id")
	}
	return handler(blockchain, stream)
}
