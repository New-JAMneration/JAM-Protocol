package ce

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// HandlePreimageRequest handles the request for a preimage of the given hash.
//
// Protocol CE143:
// Node -> Node
//
//	--> Hash
//	--> FIN
//	<-- Preimage
//	<-- FIN
//
// The request structure is:
// - Hash: 32 bytes (OpaqueHash)
//
// Total request message size: 32 bytes
func HandlePreimageRequest(bc blockchain.Blockchain, stream io.ReadWriteCloser) error {
	hashSize := 32
	hashData := make([]byte, hashSize)

	if _, err := io.ReadFull(stream, hashData); err != nil {
		return fmt.Errorf("failed to read hash data: %w", err)
	}

	hash := types.OpaqueHash{}
	copy(hash[:], hashData)

	if err := expectRemoteFIN(stream); err != nil {
		return err
	}

	preimage, err := getPreimageFromStorage(bc, hash)
	if err != nil {
		return fmt.Errorf("failed to retrieve preimage: %w", err)
	}

	if _, err := stream.Write(preimage); err != nil {
		return fmt.Errorf("failed to write preimage response: %w", err)
	}
	return stream.Close()
}

// getPreimageFromStorage retrieves the preimage from the CE database
func getPreimageFromStorage(bc blockchain.Blockchain, hash types.OpaqueHash) ([]byte, error) {
	db := DB(bc)
	preimage, err := GetKV(db, cePreimageKey(hash))
	if err != nil {
		return nil, fmt.Errorf("failed to get preimage from storage: %w", err)
	}
	if preimage == nil {
		return nil, fmt.Errorf("preimage not found for hash: %x", hash)
	}
	return preimage, nil
}

// StorePreimage stores a preimage in the CE database
func StorePreimage(bc blockchain.Blockchain, hashValue types.OpaqueHash, preimage []byte) error {
	if len(preimage) == 0 {
		return errors.New("preimage cannot be empty")
	}
	const maxPreimageSize = 100 * 1024 * 1024
	if len(preimage) > maxPreimageSize {
		return fmt.Errorf("preimage too large: %d bytes (max: %d)", len(preimage), maxPreimageSize)
	}
	preimageHash := hash.Blake2bHash(types.ByteSequence(preimage))
	if preimageHash != hashValue {
		return fmt.Errorf("preimage hash mismatch: expected %x, got %x", hashValue, preimageHash)
	}
	db := DB(bc)
	if err := PutKV(db, cePreimageKey(hashValue), preimage); err != nil {
		return fmt.Errorf("failed to store preimage: %w", err)
	}
	return nil
}

// CreatePreimageRequest creates a preimage request message (hash only; sender closes after write).
func CreatePreimageRequest(hash types.OpaqueHash) ([]byte, error) {
	request := make([]byte, 32)
	copy(request[:32], hash[:])
	return request, nil
}

func (h *DefaultCERequestHandler) encodePreimageRequest(message interface{}) ([]byte, error) {
	request, ok := message.(*CE143Payload)
	if !ok {
		return nil, fmt.Errorf("unsupported message type for PreimageRequest: %T", message)
	}

	if request == nil {
		return nil, fmt.Errorf("nil payload for PreimageRequest")
	}

	result := make([]byte, 32)
	copy(result, request.Hash[:])

	return result, nil
}

type CE143Payload struct {
	Hash     types.OpaqueHash
	Preimage []byte
}

func (p *CE143Payload) Validate() error {
	if len(p.Preimage) == 0 {
		return errors.New("preimage cannot be empty")
	}

	const maxPreimageSize = 100 * 1024 * 1024
	if len(p.Preimage) > maxPreimageSize {
		return fmt.Errorf("preimage too large: %d bytes (max: %d)", len(p.Preimage), maxPreimageSize)
	}

	preimageHash := hash.Blake2bHash(types.ByteSequence(p.Preimage))
	if preimageHash != p.Hash {
		return fmt.Errorf("preimage hash mismatch: expected %x, got %x", p.Hash, preimageHash)
	}

	return nil
}

func (p *CE143Payload) Encode() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	// Hash (32 bytes) + Preimage length (4 bytes) + Preimage
	encoded := make([]byte, 32+4+len(p.Preimage))

	copy(encoded[:32], p.Hash[:])
	binary.LittleEndian.PutUint32(encoded[32:36], uint32(len(p.Preimage)))
	copy(encoded[36:], p.Preimage)

	return encoded, nil
}

// Decode decodes bytes to CE143Payload
func (p *CE143Payload) Decode(data []byte) error {
	if len(data) < 36 {
		return fmt.Errorf("invalid data size: expected at least 36 bytes, got %d", len(data))
	}

	copy(p.Hash[:], data[:32])
	preimageLength := binary.LittleEndian.Uint32(data[32:36])

	if len(data) != int(36+preimageLength) {
		return fmt.Errorf("invalid data size: expected %d bytes, got %d", 36+preimageLength, len(data))
	}

	p.Preimage = make([]byte, preimageLength)
	copy(p.Preimage, data[36:])

	return p.Validate()
}
