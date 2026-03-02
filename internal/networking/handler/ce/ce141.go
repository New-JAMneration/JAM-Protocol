package ce

import (
	"crypto/ed25519"
	"fmt"
	"io"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
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
// When peerEd25519 is provided (e.g. from the TLS peer certificate), the signature is verified
// against the message Header Hash ++ Bitfield. When nil, signature verification is skipped (e.g. tests).
func HandleAvailabilityAssuranceDistribution(blockchain blockchain.Blockchain, stream io.ReadWriteCloser, peerEd25519 ...ed25519.PublicKey) error {
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

	if err := expectRemoteFIN(stream); err != nil {
		return err
	}

	if err := validateAvailabilityAssurance(headerHash, bitfield, signature, peerEd25519); err != nil {
		return fmt.Errorf("invalid availability assurance: %w", err)
	}

	if err := storeAvailabilityAssurance(blockchain, headerHash, bitfield, signature); err != nil {
		return fmt.Errorf("failed to store availability assurance: %w", err)
	}
	return stream.Close()
}

// validateAvailabilityAssurance validates the received availability assurance.
// When peerEd25519 is non-empty, the signature is verified using the first key (assurer's cert key).
func validateAvailabilityAssurance(headerHash types.HeaderHash, bitfield []byte, signature types.Ed25519Signature, peerEd25519 []ed25519.PublicKey) error {
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

	if len(peerEd25519) > 0 && peerEd25519[0] != nil {
		message := make([]byte, 32+len(bitfield))
		copy(message[:32], headerHash[:])
		copy(message[32:], bitfield)
		if !ed25519.Verify(peerEd25519[0], message, signature[:]) {
			return fmt.Errorf("invalid Ed25519 signature")
		}
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
func storeAvailabilityAssurance(bc blockchain.Blockchain, headerHash types.HeaderHash, bitfield []byte, signature types.Ed25519Signature) error {
	db := DB(bc)
	assuranceData := &CE141Payload{
		HeaderHash: headerHash,
		Bitfield:   bitfield,
		Signature:  signature,
	}

	encodedAssurance, err := assuranceData.Encode()
	if err != nil {
		return fmt.Errorf("failed to encode assurance data: %w", err)
	}

	if err := PutKV(db, ceAssuranceKey(headerHash), encodedAssurance); err != nil {
		return fmt.Errorf("failed to store assurance: %w", err)
	}
	if err := SAdd(db, ceAssuranceSetKey(headerHash), encodedAssurance); err != nil {
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

func (h *DefaultCERequestHandler) encodeAssuranceDistribution(message interface{}) ([]byte, error) {
	assurance, ok := message.(*CE141Payload)
	if !ok {
		return nil, fmt.Errorf("unsupported message type for AssuranceDistribution: %T", message)
	}

	if assurance == nil {
		return nil, fmt.Errorf("nil payload for AssuranceDistribution")
	}

	if err := assurance.Validate(); err != nil {
		return nil, fmt.Errorf("invalid assurance payload: %w", err)
	}

	assuranceBytes, err := assurance.Encode()
	if err != nil {
		return nil, fmt.Errorf("failed to encode assurance data: %w", err)
	}

	totalLen := len(assuranceBytes)
	result := make([]byte, 0, totalLen)

	result = append(result, assuranceBytes...)

	return result, nil
}
