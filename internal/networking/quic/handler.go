package quic

import (
	"encoding/json"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// StreamHandler interface combines UPHandler and CEHandler for complete stream handling
type StreamHandler interface {
	UPHandler // Unique Persistent
	CEHandler // Common Ephemeral
}

// UPHandler interface for handling unique persistent streams
type UPHandler interface {
	EncodeMessage(kind string, message interface{}) ([]byte, error)
}

type CEHandler interface {
	HandleStream(stream *Stream) error
}

type DefaultUPHandler struct {
	encoder *types.Encoder
}

func NewDefaultUPHandler() *DefaultUPHandler {
	return &DefaultUPHandler{
		encoder: types.NewEncoder(),
	}
}

func (h *DefaultUPHandler) EncodeMessage(kind string, message interface{}) ([]byte, error) {
	switch kind {
	case "block":
		if block, ok := message.(*types.Block); ok {
			return h.encoder.Encode(block)
		}
	case "header":
		if header, ok := message.(*types.Header); ok {
			return h.encoder.Encode(header)
		}
	case "work_package":
		if wp, ok := message.(*types.WorkPackage); ok {
			return h.encoder.Encode(wp)
		}
	case "json":
		return json.Marshal(message)
	default:
		if encodable, ok := message.(types.Encodable); ok {
			return h.encoder.Encode(encodable)
		}
		return json.Marshal(message)
	}

	return nil, fmt.Errorf("unsupported message kind: %s", kind)
}

type DefaultCEHandler struct {
	blockchain blockchain.Blockchain
	handlers   map[uint8]func(blockchain.Blockchain, *Stream) error
}

func NewDefaultCEHandler(bc blockchain.Blockchain) *DefaultCEHandler {
	return &DefaultCEHandler{
		blockchain: bc,
		handlers:   make(map[uint8]func(blockchain.Blockchain, *Stream) error),
	}
}

func (h *DefaultCEHandler) HandleStream(stream *Stream) error {
	protocolID := make([]byte, 1)
	if _, err := stream.Read(protocolID); err != nil {
		return fmt.Errorf("failed to read protocol ID: %w", err)
	}

	handler, exists := h.handlers[protocolID[0]]
	if !exists {
		return fmt.Errorf("unsupported protocol ID: %d", protocolID[0])
	}

	return handler(h.blockchain, stream)
}

func (h *DefaultCEHandler) RegisterCEHandler(protoID uint8, handlerFunc func(blockchain.Blockchain, *Stream) error) {
	h.handlers[protoID] = handlerFunc
}

type DefaultStreamHandler struct {
	*DefaultUPHandler
	*DefaultCEHandler
}

func NewDefaultStreamHandler(bc blockchain.Blockchain) *DefaultStreamHandler {
	return &DefaultStreamHandler{
		DefaultUPHandler: NewDefaultUPHandler(),
		DefaultCEHandler: NewDefaultCEHandler(bc),
	}
}
