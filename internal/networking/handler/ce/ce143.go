package ce

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
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
func HandlePreimageRequest(blockchain blockchain.Blockchain, stream io.ReadWriteCloser) error {
	hashSize := 32
	hashData := make([]byte, hashSize)

	if _, err := io.ReadFull(stream, hashData); err != nil {
		return fmt.Errorf("failed to read hash data: %w", err)
	}

	hash := types.OpaqueHash{}
	copy(hash[:], hashData)

	finBuf := make([]byte, 3)
	if _, err := io.ReadFull(stream, finBuf); err != nil {
		return fmt.Errorf("failed to read FIN: %w", err)
	}
	if string(finBuf) != "FIN" {
		return errors.New("request does not end with FIN")
	}

	preimage, err := getPreimageFromStorage(hash)
	if err != nil {
		return fmt.Errorf("failed to retrieve preimage: %w", err)
	}

	if _, err := stream.Write(preimage); err != nil {
		return fmt.Errorf("failed to write preimage response: %w", err)
	}

	if _, err := stream.Write([]byte("FIN")); err != nil {
		return fmt.Errorf("failed to write FIN response: %w", err)
	}

	return stream.Close()
}

// getPreimageFromStorage retrieves the preimage from the preimage database
func getPreimageFromStorage(hash types.OpaqueHash) ([]byte, error) {
	redisBackend, err := store.GetRedisBackend()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis backend: %w", err)
	}

	hashHex := hex.EncodeToString(hash[:])
	key := fmt.Sprintf("preimage:%s", hashHex)

	client := redisBackend.GetClient()
	preimage, err := client.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get preimage from Redis: %w", err)
	}

	if preimage == nil {
		return nil, fmt.Errorf("preimage not found for hash: %x", hash)
	}

	return preimage, nil
}

// StorePreimage stores a preimage in the preimage database
func StorePreimage(hash types.OpaqueHash, preimage []byte) error {
	if len(preimage) == 0 {
		return errors.New("preimage cannot be empty")
	}

	const maxPreimageSize = 100 * 1024 * 1024
	if len(preimage) > maxPreimageSize {
		return fmt.Errorf("preimage too large: %d bytes (max: %d)", len(preimage), maxPreimageSize)
	}

	redisBackend, err := store.GetRedisBackend()
	if err != nil {
		return fmt.Errorf("failed to get Redis backend: %w", err)
	}

	hashHex := hex.EncodeToString(hash[:])
	key := fmt.Sprintf("preimage:%s", hashHex)

	client := redisBackend.GetClient()
	err = client.Put(key, preimage)
	if err != nil {
		return fmt.Errorf("failed to store preimage in Redis: %w", err)
	}

	return nil
}

// CreatePreimageRequest creates a preimage request message
func CreatePreimageRequest(hash types.OpaqueHash) ([]byte, error) {
	request := make([]byte, 32+3) // Hash + FIN

	copy(request[:32], hash[:])
	copy(request[32:35], []byte("FIN"))

	return request, nil
}

// GetPreimage retrieves a preimage from Redis storage
func GetPreimage(hash types.OpaqueHash) ([]byte, error) {
	return getPreimageFromStorage(hash)
}

// DeletePreimage removes a preimage from Redis storage
func DeletePreimage(hash types.OpaqueHash) error {
	redisBackend, err := store.GetRedisBackend()
	if err != nil {
		return fmt.Errorf("failed to get Redis backend: %w", err)
	}

	hashHex := hex.EncodeToString(hash[:])
	key := fmt.Sprintf("preimage:%s", hashHex)

	client := redisBackend.GetClient()
	err = client.Delete(key)
	if err != nil {
		return fmt.Errorf("failed to delete preimage from Redis: %w", err)
	}

	return nil
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
