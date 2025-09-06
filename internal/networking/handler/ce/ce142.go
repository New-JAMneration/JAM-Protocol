package ce

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// HandlePreimageAnnouncement handles the announcement of possession of a requested preimage.
func HandlePreimageAnnouncement(blockchain blockchain.Blockchain, stream *quic.Stream) error {
	// Service ID (4 bytes) + Hash (32 bytes) + Preimage Length (4 bytes)
	announcementSize := 4 + 32 + 4
	announcementData := make([]byte, announcementSize)

	if err := stream.ReadFull(announcementData); err != nil {
		return fmt.Errorf("failed to read announcement data: %w", err)
	}

	// Parse the announcement components
	serviceID := types.ServiceId(binary.LittleEndian.Uint32(announcementData[:4]))

	hash := types.OpaqueHash{}
	copy(hash[:], announcementData[4:36])

	preimageLength := types.U32(binary.LittleEndian.Uint32(announcementData[36:40]))

	finBuf := make([]byte, 3)
	if err := stream.ReadFull(finBuf); err != nil {
		return fmt.Errorf("failed to read FIN: %w", err)
	}
	if string(finBuf) != "FIN" {
		return errors.New("request does not end with FIN")
	}

	if err := validatePreimageAnnouncement(serviceID, hash, preimageLength); err != nil {
		return fmt.Errorf("invalid preimage announcement: %w", err)
	}

	if err := storePreimageAnnouncement(serviceID, hash, preimageLength); err != nil {
		return fmt.Errorf("failed to store preimage announcement: %w", err)
	}

	if _, err := stream.Write([]byte("FIN")); err != nil {
		return fmt.Errorf("failed to write FIN response: %w", err)
	}

	return stream.Close()
}

// validatePreimageAnnouncement validates the received preimage announcement
func validatePreimageAnnouncement(serviceID types.ServiceId, hash types.OpaqueHash, preimageLength types.U32) error {
	if preimageLength == 0 {
		return errors.New("preimage length cannot be zero")
	}

	const maxPreimageSize = 100 * 1024 * 1024
	if preimageLength > types.U32(maxPreimageSize) {
		return fmt.Errorf("preimage length too large: %d bytes (max: %d)", preimageLength, maxPreimageSize)
	}

	return nil
}

// storePreimageAnnouncement stores the received preimage announcement
func storePreimageAnnouncement(serviceID types.ServiceId, hash types.OpaqueHash, preimageLength types.U32) error {
	redisBackend, err := store.GetRedisBackend()
	if err != nil {
		return fmt.Errorf("failed to get Redis backend: %w", err)
	}

	announcementData := &CE142Payload{
		ServiceID:      serviceID,
		Hash:           hash,
		PreimageLength: preimageLength,
	}

	encodedAnnouncement, err := announcementData.Encode()
	if err != nil {
		return fmt.Errorf("failed to encode announcement data: %w", err)
	}

	hashHex := hex.EncodeToString(hash[:])
	key := fmt.Sprintf("preimage_announcement:%s", hashHex)

	client := redisBackend.GetClient()
	err = client.Put(key, encodedAnnouncement)
	if err != nil {
		return fmt.Errorf("failed to store announcement in Redis: %w", err)
	}

	// Store in a set for service-based lookups
	serviceKey := fmt.Sprintf("service_preimage_announcements:%d", serviceID)
	err = client.SAdd(serviceKey, encodedAnnouncement)
	if err != nil {
		return fmt.Errorf("failed to add announcement to service set: %w", err)
	}

	return nil
}

// CreatePreimageAnnouncement creates a preimage announcement message
func CreatePreimageAnnouncement(
	serviceID types.ServiceId,
	hash types.OpaqueHash,
	preimageLength types.U32,
) ([]byte, error) {
	// Validate preimage length
	if preimageLength == 0 {
		return nil, errors.New("preimage length cannot be zero")
	}

	const maxPreimageSize = 100 * 1024 * 1024
	if preimageLength > types.U32(maxPreimageSize) {
		return nil, fmt.Errorf("preimage length too large: %d bytes (max: %d)", preimageLength, maxPreimageSize)
	}

	announcement := make([]byte, 4+32+4)

	binary.LittleEndian.PutUint32(announcement[:4], uint32(serviceID))

	copy(announcement[4:36], hash[:])

	binary.LittleEndian.PutUint32(announcement[36:40], uint32(preimageLength))

	return announcement, nil
}

type CE142Payload struct {
	ServiceID      types.ServiceId
	Hash           types.OpaqueHash
	PreimageLength types.U32
}

func (p *CE142Payload) Validate() error {
	if p.PreimageLength == 0 {
		return errors.New("preimage length cannot be zero")
	}

	const maxPreimageSize = 100 * 1024 * 1024
	if p.PreimageLength > types.U32(maxPreimageSize) {
		return fmt.Errorf("preimage length too large: %d bytes (max: %d)", p.PreimageLength, maxPreimageSize)
	}

	return nil
}

func (p *CE142Payload) Encode() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	encoded := make([]byte, 4+32+4)

	binary.LittleEndian.PutUint32(encoded[:4], uint32(p.ServiceID))

	copy(encoded[4:36], p.Hash[:])

	binary.LittleEndian.PutUint32(encoded[36:40], uint32(p.PreimageLength))

	return encoded, nil
}

// Decode decodes bytes to CE142Payload
func (p *CE142Payload) Decode(data []byte) error {
	expectedSize := 4 + 32 + 4

	if len(data) != expectedSize {
		return fmt.Errorf("invalid data size: expected %d, got %d", expectedSize, len(data))
	}

	p.ServiceID = types.ServiceId(binary.LittleEndian.Uint32(data[:4]))

	copy(p.Hash[:], data[4:36])

	p.PreimageLength = types.U32(binary.LittleEndian.Uint32(data[36:40]))

	return p.Validate()
}

func (h *DefaultCERequestHandler) encodePreimageAnnouncement(message interface{}) ([]byte, error) {
	announcement, ok := message.(*CE142Payload)
	if !ok {
		return nil, fmt.Errorf("unsupported message type for PreimageAnnouncement: %T", message)
	}

	if announcement == nil {
		return nil, fmt.Errorf("nil payload for PreimageAnnouncement")
	}

	if err := announcement.Validate(); err != nil {
		return nil, fmt.Errorf("invalid announcement payload: %w", err)
	}

	announcementBytes, err := announcement.Encode()
	if err != nil {
		return nil, fmt.Errorf("failed to encode announcement data: %w", err)
	}

	totalLen := len(announcementBytes)
	result := make([]byte, 0, totalLen)

	result = append(result, announcementBytes...)

	return result, nil
}
