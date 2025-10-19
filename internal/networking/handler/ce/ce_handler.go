package ce

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
)

type Role uint8

const (
	Assurer Role = iota + 1
	Guarantor
	Auditor
	Validator
	Builder
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

// HandleStream reads a framed request from the given stream, checks the protocol ID,
// and dispatches to the registered handler. The stream is expected to be framed
// with a 4-byte little-endian length prefix.
func (h *CEHandler) HandleStream(blockchain blockchain.Blockchain, stream *quic.Stream) error {
	// Read the 4-byte length prefix.
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(stream, lenBuf); err != nil {
		return err
	}
	msgLen := binary.LittleEndian.Uint32(lenBuf)
	payload := make([]byte, msgLen)
	if _, err := io.ReadFull(stream, payload); err != nil {
		return err
	}

	// The first byte is the protocol ID.
	if len(payload) < 1 {
		return errors.New("payload too short, missing protocol id")
	}
	protoID := payload[0]

	handler, ok := h.handlers[protoID]
	if !ok {
		return errors.New("unsupported CE request protocol id")
	}

	// Dispatch to the handler.
	return handler(blockchain, stream)
}
