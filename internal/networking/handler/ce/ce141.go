package ce

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// HandleAvailabilityAssuranceDistribution handles the distribution of availability assurances
// from assurers to validators approximately 2 seconds before each slot.
//
// Protocol CE141:
// Assurer -> Validator
//
//	--> Assurance (Header Hash (Anchor) ++ Bitfield ++ Ed25519 Signature)
//	--> FIN
//	<-- FIN
//
// The bitfield is [u8; ceil(C / 8)] where C is the total number of cores.
// One bit per core indicates availability.
//
// The assurance structure is:
// - Header Hash (Anchor): 32 bytes
// - Bitfield: ceil(C / 8) bytes (where C = CoresCount)
// - Ed25519 Signature: 64 bytes
//
// Total message size: 32 + ceil(C/8) + 64 bytes
func HandleAvailabilityAssuranceDistribution(blockchain blockchain.Blockchain, stream io.ReadWriteCloser) error {
	bitfieldSize := (types.CoresCount + 7) / 8 // ceil(C / 8)

	// Header Hash (32 bytes) + Bitfield (ceil(C/8) bytes) + Ed25519 Signature (64 bytes)
	assuranceSize := 32 + bitfieldSize + 64
	assuranceData := make([]byte, assuranceSize)

	if _, err := io.ReadFull(stream, assuranceData); err != nil {
		return fmt.Errorf("failed to read assurance data: %w", err)
	}

	// Parse the assurance components
	headerHash := types.HeaderHash{}
	copy(headerHash[:], assuranceData[:32])

	bitfield := make([]byte, bitfieldSize)
	copy(bitfield, assuranceData[32:32+bitfieldSize])

	signature := types.Ed25519Signature{}
	copy(signature[:], assuranceData[32+bitfieldSize:])

	finBuf := make([]byte, 3)
	if _, err := io.ReadFull(stream, finBuf); err != nil {
		return fmt.Errorf("failed to read FIN: %w", err)
	}
	if string(finBuf) != "FIN" {
		return errors.New("request does not end with FIN")
	}

	if err := validateAvailabilityAssurance(headerHash, bitfield, signature); err != nil {
		return fmt.Errorf("invalid availability assurance: %w", err)
	}

	if err := storeAvailabilityAssurance(headerHash, bitfield, signature); err != nil {
		return fmt.Errorf("failed to store availability assurance: %w", err)
	}

	if _, err := stream.Write([]byte("FIN")); err != nil {
		return fmt.Errorf("failed to write FIN response: %w", err)
	}

	return stream.Close()
}

// validateAvailabilityAssurance validates the received availability assurance
func validateAvailabilityAssurance(headerHash types.HeaderHash, bitfield []byte, signature types.Ed25519Signature) error {
	// Validate bitfield size
	expectedBitfieldSize := (types.CoresCount + 7) / 8
	if len(bitfield) != expectedBitfieldSize {
		return fmt.Errorf("invalid bitfield size: expected %d, got %d", expectedBitfieldSize, len(bitfield))
	}

	if len(signature) != 64 {
		return fmt.Errorf("invalid signature length: expected 64, got %d", len(signature))
	}

	if err := validateBitfieldFormat(bitfield); err != nil {
		return fmt.Errorf("invalid bitfield format: %w", err)
	}

	return nil
}

// validateBitfieldFormat validates the bitfield format
func validateBitfieldFormat(bitfield []byte) error {
	// Check that unused bits in the last byte are zero
	lastByteIndex := len(bitfield) - 1
	lastByte := bitfield[lastByteIndex]

	// Calculate how many bits are actually used in the last byte
	usedBitsInLastByte := types.CoresCount % 8
	if usedBitsInLastByte == 0 {
		usedBitsInLastByte = 8
	}

	// Check that unused bits are zero
	// For example, with 2 cores, usedBitsInLastByte = 2, so unusedBitsMask = 0xFC (11111100)
	// This means bits 2-7 should be zero
	unusedBitsMask := byte(0xFF << usedBitsInLastByte)
	if (lastByte & unusedBitsMask) != 0 {
		return fmt.Errorf("unused bits in last byte are not zero: last byte = %02x, used bits = %d, unused mask = %02x", lastByte, usedBitsInLastByte, unusedBitsMask)
	}

	return nil
}

// storeAvailabilityAssurance stores the received availability assurance
func storeAvailabilityAssurance(headerHash types.HeaderHash, bitfield []byte, signature types.Ed25519Signature) error {
	redisBackend, err := store.GetRedisBackend()
	if err != nil {
		return fmt.Errorf("failed to get Redis backend: %w", err)
	}

	assuranceData := &CE141Payload{
		HeaderHash: headerHash,
		Bitfield:   bitfield,
		Signature:  signature,
	}

	encodedAssurance, err := assuranceData.Encode()
	if err != nil {
		return fmt.Errorf("failed to encode assurance data: %w", err)
	}

	headerHashHex := hex.EncodeToString(headerHash[:])
	key := fmt.Sprintf("availability_assurance:%s", headerHashHex)

	client := redisBackend.GetClient()
	err = client.Put(key, encodedAssurance)
	if err != nil {
		return fmt.Errorf("failed to store assurance in Redis: %w", err)
	}

	setKey := fmt.Sprintf("availability_assurances_set:%s", headerHashHex)
	err = client.SAdd(setKey, encodedAssurance)
	if err != nil {
		return fmt.Errorf("failed to add assurance to set: %w", err)
	}

	return nil
}

// CreateAvailabilityAssurance creates an availability assurance message for distribution
func CreateAvailabilityAssurance(
	headerHash types.HeaderHash,
	bitfield []byte,
	privateKey ed25519.PrivateKey,
) ([]byte, error) {
	// Validate bitfield size
	expectedBitfieldSize := (types.CoresCount + 7) / 8
	if len(bitfield) != expectedBitfieldSize {
		return nil, fmt.Errorf("invalid bitfield size: expected %d, got %d", expectedBitfieldSize, len(bitfield))
	}

	// Validate bitfield format
	if err := validateBitfieldFormat(bitfield); err != nil {
		return nil, fmt.Errorf("invalid bitfield format: %w", err)
	}

	// Create the message to sign: Header Hash ++ Bitfield
	message := make([]byte, 32+len(bitfield))
	copy(message[:32], headerHash[:])
	copy(message[32:], bitfield)

	signature := ed25519.Sign(privateKey, message)

	assurance := make([]byte, 32+len(bitfield)+64)
	copy(assurance[:32], headerHash[:])
	copy(assurance[32:32+len(bitfield)], bitfield)
	copy(assurance[32+len(bitfield):], signature)

	return assurance, nil
}

// VerifyAvailabilityAssurance verifies an availability assurance signature
func VerifyAvailabilityAssurance(
	headerHash types.HeaderHash,
	bitfield []byte,
	signature types.Ed25519Signature,
	publicKey types.Ed25519Public,
) bool {
	message := make([]byte, 32+len(bitfield))
	copy(message[:32], headerHash[:])
	copy(message[32:], bitfield)

	return ed25519.Verify(publicKey[:], message, signature[:])
}

// CreateBitfieldFromCoreAvailability creates a bitfield from core availability information
// Each bit represents whether a core is available (1) or not (0)
func CreateBitfieldFromCoreAvailability(coreAvailability []bool) ([]byte, error) {
	if len(coreAvailability) != types.CoresCount {
		return nil, fmt.Errorf("core availability array must have exactly %d elements", types.CoresCount)
	}

	bitfieldSize := (types.CoresCount + 7) / 8
	bitfield := make([]byte, bitfieldSize)

	for i, available := range coreAvailability {
		byteIndex := i / 8
		bitIndex := i % 8

		if available {
			bitfield[byteIndex] |= (1 << bitIndex)
		}
	}

	return bitfield, nil
}

// ParseBitfieldToCoreAvailability parses a bitfield back to core availability information
func ParseBitfieldToCoreAvailability(bitfield []byte) ([]bool, error) {
	expectedBitfieldSize := (types.CoresCount + 7) / 8
	if len(bitfield) != expectedBitfieldSize {
		return nil, fmt.Errorf("invalid bitfield size: expected %d, got %d", expectedBitfieldSize, len(bitfield))
	}

	coreAvailability := make([]bool, types.CoresCount)

	for i := 0; i < types.CoresCount; i++ {
		byteIndex := i / 8
		bitIndex := i % 8

		coreAvailability[i] = (bitfield[byteIndex] & (1 << bitIndex)) != 0
	}

	return coreAvailability, nil
}

// GetAvailabilityAssurance retrieves an availability assurance from Redis storage
func GetAvailabilityAssurance(headerHash types.HeaderHash) (*CE141Payload, error) {
	// Get Redis backend
	redisBackend, err := store.GetRedisBackend()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis backend: %w", err)
	}

	// Create Redis key for the assurance
	headerHashHex := hex.EncodeToString(headerHash[:])
	key := fmt.Sprintf("availability_assurance:%s", headerHashHex)

	// Retrieve the assurance from Redis
	client := redisBackend.GetClient()
	encodedAssurance, err := client.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get assurance from Redis: %w", err)
	}

	if encodedAssurance == nil {
		return nil, fmt.Errorf("availability assurance not found for header hash: %x", headerHash)
	}

	// Decode the assurance data
	assuranceData := &CE141Payload{}
	err = assuranceData.Decode(encodedAssurance)
	if err != nil {
		return nil, fmt.Errorf("failed to decode assurance data: %w", err)
	}

	return assuranceData, nil
}

// GetAllAvailabilityAssurancesForHeader retrieves all availability assurances for a given header hash
func GetAllAvailabilityAssurancesForHeader(headerHash types.HeaderHash) ([]*CE141Payload, error) {
	// Get Redis backend
	redisBackend, err := store.GetRedisBackend()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis backend: %w", err)
	}

	// Create Redis set key
	headerHashHex := hex.EncodeToString(headerHash[:])
	setKey := fmt.Sprintf("availability_assurances_set:%s", headerHashHex)

	// Get all members from the set
	client := redisBackend.GetClient()
	encodedAssurances, err := client.SMembers(setKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get assurances set from Redis: %w", err)
	}

	// Decode each assurance
	var assurances []*CE141Payload
	for _, encodedAssurance := range encodedAssurances {
		assuranceData := &CE141Payload{}
		err := assuranceData.Decode(encodedAssurance)
		if err != nil {
			return nil, fmt.Errorf("failed to decode assurance data: %w", err)
		}
		assurances = append(assurances, assuranceData)
	}

	return assurances, nil
}

// DeleteAvailabilityAssurance removes an availability assurance from Redis storage
func DeleteAvailabilityAssurance(headerHash types.HeaderHash) error {
	// Get Redis backend
	redisBackend, err := store.GetRedisBackend()
	if err != nil {
		return fmt.Errorf("failed to get Redis backend: %w", err)
	}

	// Create Redis keys
	headerHashHex := hex.EncodeToString(headerHash[:])
	key := fmt.Sprintf("availability_assurance:%s", headerHashHex)
	setKey := fmt.Sprintf("availability_assurances_set:%s", headerHashHex)

	// Get the encoded assurance to remove from set
	client := redisBackend.GetClient()
	encodedAssurance, err := client.Get(key)
	if err != nil {
		return fmt.Errorf("failed to get assurance from Redis: %w", err)
	}

	if encodedAssurance != nil {
		// Remove from set
		err = client.SRem(setKey, encodedAssurance)
		if err != nil {
			return fmt.Errorf("failed to remove assurance from set: %w", err)
		}
	}

	// Delete the main key
	err = client.Delete(key)
	if err != nil {
		return fmt.Errorf("failed to delete assurance from Redis: %w", err)
	}

	return nil
}

// CE141Payload represents an availability assurance message
type CE141Payload struct {
	HeaderHash types.HeaderHash
	Bitfield   []byte // ceil(C/8) bytes representing core availability
	Signature  types.Ed25519Signature
}

// Validate validates the CE141Payload
func (p *CE141Payload) Validate() error {
	expectedBitfieldSize := (types.CoresCount + 7) / 8
	if len(p.Bitfield) != expectedBitfieldSize {
		return fmt.Errorf("invalid bitfield size: expected %d, got %d", expectedBitfieldSize, len(p.Bitfield))
	}

	if len(p.Signature) != 64 {
		return fmt.Errorf("invalid signature length: expected 64, got %d", len(p.Signature))
	}

	if err := validateBitfieldFormat(p.Bitfield); err != nil {
		return fmt.Errorf("invalid bitfield format: %w", err)
	}

	return nil
}

// Encode encodes the CE141Payload to bytes
func (p *CE141Payload) Encode() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	encoded := make([]byte, 32+len(p.Bitfield)+64)
	copy(encoded[:32], p.HeaderHash[:])
	copy(encoded[32:32+len(p.Bitfield)], p.Bitfield)
	copy(encoded[32+len(p.Bitfield):], p.Signature[:])

	return encoded, nil
}

// Decode decodes bytes to CE141Payload
func (p *CE141Payload) Decode(data []byte) error {
	expectedBitfieldSize := (types.CoresCount + 7) / 8
	expectedSize := 32 + expectedBitfieldSize + 64

	if len(data) != expectedSize {
		return fmt.Errorf("invalid data size: expected %d, got %d", expectedSize, len(data))
	}

	copy(p.HeaderHash[:], data[:32])

	p.Bitfield = make([]byte, expectedBitfieldSize)
	copy(p.Bitfield, data[32:32+expectedBitfieldSize])

	copy(p.Signature[:], data[32+expectedBitfieldSize:])

	return p.Validate()
}
