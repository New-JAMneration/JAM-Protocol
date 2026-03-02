package ce

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// CE129Payload represents a state key range request message.
type CE129Payload struct {
	HeaderHash types.HeaderHash
	KeyStart   types.StateKey
	KeyEnd     types.StateKey
	MaxSize    uint32
}

// HandleStateRequest handles a CE129 state key range request.
func HandleStateRequest(blockchain blockchain.Blockchain, req CE129Payload, stream *quic.Stream) error {
	stateValues, err := blockchain.GetStateRange(req.HeaderHash, req.KeyStart, req.KeyEnd, req.MaxSize)
	if err != nil {
		return fmt.Errorf("failed to get state range: %w", err)
	}

	// Get boundary nodes for the key range
	boundaryNodes, err := blockchain.GetBoundaryNodes(req.HeaderHash, req.KeyStart, req.KeyEnd, req.MaxSize)
	if err != nil {
		return fmt.Errorf("failed to get boundary nodes: %w", err)
	}

	encoder := types.NewEncoder()

	// Build boundary message: 4-byte count + for each node 4-byte length + encoded node
	numBoundary := uint32(len(boundaryNodes))
	boundaryBlob := make([]byte, 0, 4+len(boundaryNodes)*(4+256))
	boundaryBlob = append(boundaryBlob, encodeLE32(numBoundary)...)
	for _, node := range boundaryNodes {
		encodedNode, err := encoder.Encode(&node)
		if err != nil {
			return fmt.Errorf("failed to encode boundary node: %w", err)
		}
		boundaryBlob = append(boundaryBlob, encodeLE32(uint32(len(encodedNode)))...)
		boundaryBlob = append(boundaryBlob, encodedNode...)
	}

	// Build key/values message: 4-byte count + for each 4-byte length + encoded value
	numValues := uint32(len(stateValues))
	keyValuesBlob := make([]byte, 0, 4+len(stateValues)*(4+256))
	keyValuesBlob = append(keyValuesBlob, encodeLE32(numValues)...)
	for _, stateVal := range stateValues {
		encodedVal, err := encoder.Encode(&stateVal)
		if err != nil {
			return fmt.Errorf("failed to encode state value: %w", err)
		}
		keyValuesBlob = append(keyValuesBlob, encodeLE32(uint32(len(encodedVal)))...)
		keyValuesBlob = append(keyValuesBlob, encodedVal...)
	}

	if err := stream.WriteMessage(boundaryBlob); err != nil {
		return fmt.Errorf("failed to write boundary nodes message: %w", err)
	}
	if err := stream.WriteMessage(keyValuesBlob); err != nil {
		return fmt.Errorf("failed to write key/values message: %w", err)
	}
	return stream.Close()
}

// HandleStateRequestStream reads the CE129 request from the stream and invokes the handler.
func HandleStateRequestStream(blockchain blockchain.Blockchain, stream *quic.Stream) error {
	reqPayload, err := stream.ReadMessage()
	if err != nil {
		return err
	}

	// The payload should be 32 (header hash) + 31 (key start) + 31 (key end) + 4 (max size) = 98 bytes
	if len(reqPayload) < 98 {
		return errors.New("invalid state request length")
	}

	var req CE129Payload
	copy(req.HeaderHash[:], reqPayload[:32])
	copy(req.KeyStart[:], reqPayload[32:63])
	copy(req.KeyEnd[:], reqPayload[63:94])
	req.MaxSize = binary.LittleEndian.Uint32(reqPayload[94:98])

	return HandleStateRequest(blockchain, req, stream)
}

func (h *DefaultCERequestHandler) encodeStateRequest(message interface{}) ([]byte, error) {
	stateReq, ok := message.(*CE129Payload)
	if !ok {
		return nil, fmt.Errorf("unsupported message type for StateRequest: %T", message)
	}

	maxSize := stateReq.MaxSize
	maxSizeBytes := []byte{
		byte(maxSize),
		byte(maxSize >> 8),
		byte(maxSize >> 16),
		byte(maxSize >> 24),
	}
	result := make([]byte, 0, 98)
	result = append(result, stateReq.HeaderHash[:]...)
	result = append(result, stateReq.KeyStart[:]...)
	result = append(result, stateReq.KeyEnd[:]...)
	result = append(result, maxSizeBytes...)

	return result, nil
}
