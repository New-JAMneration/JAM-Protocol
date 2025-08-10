package ce

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

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
		lengthBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(lengthBytes, uint32(len(encodedNode)))
		if _, err := stream.Write(lengthBytes); err != nil {
			return fmt.Errorf("failed to write boundary node length: %w", err)
		}
		if _, err := stream.Write(encodedNode); err != nil {
			return fmt.Errorf("failed to write encoded boundary node: %w", err)
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

		lengthBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(lengthBytes, uint32(len(encodedVal)))
		if _, err := stream.Write(lengthBytes); err != nil {
			return fmt.Errorf("failed to write value length: %w", err)
		}

		if _, err := stream.Write(encodedVal); err != nil {
			return fmt.Errorf("failed to write encoded value: %w", err)
		}
	}

	return stream.Close()
}

// HandleStateRequestStream reads the CE129 request from the stream and invokes the handler.
func HandleStateRequestStream(blockchain blockchain.Blockchain, stream *quic.Stream) error {
	// The quic.DefaultCEHandler has already read the protocol ID (1 byte).
	// Now we need to read the remaining payload from the stream.
	reqPayload, err := io.ReadAll(stream)
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
