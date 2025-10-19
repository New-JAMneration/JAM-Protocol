package ce

import (
	"encoding/binary"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// CE129Payload represents a state key range request message.
// client → server
type CE129Payload struct {
	HeaderHash types.HeaderHash
	KeyStart   types.StateKey
	KeyEnd     types.StateKey
	MaxSize    uint32
}

// CE129Response: server → client
type CE129Response struct {
	BoundaryNodes []types.ByteSequence
	Pairs         []types.StateKeyVal
}

// HandleStateRequest handles a CE129 state key range request.
// [Node -> Node]
//
// [TODO-Validation]
// 1. Ensure [blockchain.GetStateRange] returns a sorted state range.
// 2. Implement [blockchain.GetBoundaryNodes]. The detail of the function (https://github.com/zdave-parity/jam-np/blob/main/simple.md#ce-129-state-request). This function should checks wheather the given start key is present in the state trie, otherwise it should takes another approach.
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

	numBoundary := uint32(len(boundaryNodes))
	numBoundaryBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(numBoundaryBytes, numBoundary)
	if _, err := stream.Write(numBoundaryBytes); err != nil {
		return fmt.Errorf("failed to write number of boundary nodes: %w", err)
	}

	// Write each boundary node (length-prefixed encoding)
	for _, node := range boundaryNodes {
		encodedNode, err := encoder.Encode(&node)
		if err != nil {
			return fmt.Errorf("failed to encode boundary node: %w", err)
		}
		if err := stream.WriteMessage(encodedNode); err != nil {
			return fmt.Errorf("failed to write boundary node length: %w", err)
		}
	}

	numValues := uint32(len(stateValues))
	numValuesBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(numValuesBytes, numValues)
	if _, err := stream.Write(numValuesBytes); err != nil {
		return fmt.Errorf("failed to write number of values: %w", err)
	}

	for _, stateVal := range stateValues {
		encodedVal, err := encoder.Encode(&stateVal)
		if err != nil {
			return fmt.Errorf("failed to encode state value: %w", err)
		}
		if err := stream.WriteMessage(encodedVal); err != nil {
			return fmt.Errorf("failed to write value length: %w", err)
		}
	}

	return stream.Close()
}

// HandleStateRequestStream reads the CE129 request from the stream and invokes the handler.
func HandleStateRequestStream(blockchain blockchain.Blockchain, stream *quic.Stream) error {
	payload := make([]byte, 98)
	err := stream.ReadFull(payload)
	if err != nil {
		return err
	}

	// The payload should be 32 (header hash) + 31 (key start) + 31 (key end) + 4 (max size) = 98 bytes

	var req CE129Payload
	copy(req.HeaderHash[:], payload[:32])
	copy(req.KeyStart[:], payload[32:63])
	copy(req.KeyEnd[:], payload[63:94])
	req.MaxSize = binary.LittleEndian.Uint32(payload[94:98])

	return HandleStateRequest(blockchain, req, stream)
}

func (h *DefaultCERequestHandler) encodeStateRequest(message interface{}) ([]byte, error) {
	stateReq, ok := message.(*CE129Payload)
	if !ok {
		return nil, fmt.Errorf("unsupported message type for StateRequest: %T", message)
	}

	encoder := types.NewEncoder()

	writeRaw := func(b []byte) error {
		for _, v := range b {
			if err := encoder.WriteByte(v); err != nil {
				return err
			}
		}
		return nil
	}

	if err := writeRaw(stateReq.HeaderHash[:]); err != nil {
		return nil, fmt.Errorf("failed to encode HeaderHash: %w", err)
	}
	if err := writeRaw(stateReq.KeyStart[:]); err != nil {
		return nil, fmt.Errorf("failed to encode KeyStart: %w", err)
	}
	if err := writeRaw(stateReq.KeyEnd[:]); err != nil {
		return nil, fmt.Errorf("failed to encode KeyEnd: %w", err)
	}
	maxSize := stateReq.MaxSize
	maxSizeBytes := []byte{
		byte(maxSize),
		byte(maxSize >> 8),
		byte(maxSize >> 16),
		byte(maxSize >> 24),
	}
	if err := writeRaw(maxSizeBytes); err != nil {
		return nil, fmt.Errorf("failed to encode MaxSize: %w", err)
	}

	result := make([]byte, 0, 98)
	result = append(result, stateReq.HeaderHash[:]...)
	result = append(result, stateReq.KeyStart[:]...)
	result = append(result, stateReq.KeyEnd[:]...)
	result = append(result, maxSizeBytes...)

	return result, nil
}
