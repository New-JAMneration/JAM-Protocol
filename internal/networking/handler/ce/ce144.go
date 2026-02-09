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

// HandleAuditAnnouncement handles the announcement of requirement to audit.
// Auditors of a block (defined to be the prior validator set) should, at the beginning
// of each tranche, broadcast an announcement to all other such auditors specifying
// which work-reports they intend to audit.
//
// Protocol CE144:
// Auditor -> Auditor
//
//	--> Header Hash ++ Tranche ++ Announcement
//	--> Evidence
//	--> FIN
//	<-- FIN
//
// The transmission format includes:
// - Header Hash: 32 bytes (OpaqueHash)
// - Tranche: 1 byte (u8)
// - Announcement: len++[Core Index ++ Work-Report Hash] ++ Ed25519 Signature
// - Evidence: depends on tranche (Bandersnatch signature for tranche 0, more complex for others)
func HandleAuditAnnouncement(blockchain blockchain.Blockchain, stream io.ReadWriteCloser) error {
	headerHash := types.OpaqueHash{}
	if _, err := io.ReadFull(stream, headerHash[:]); err != nil {
		return fmt.Errorf("failed to read header hash: %w", err)
	}

	trancheBuf := make([]byte, 1)
	if _, err := io.ReadFull(stream, trancheBuf); err != nil {
		return fmt.Errorf("failed to read tranche: %w", err)
	}
	tranche := uint8(trancheBuf[0])

	announcement, err := readAnnouncement(stream)
	if err != nil {
		return fmt.Errorf("failed to read announcement: %w", err)
	}
	evidence, err := readEvidence(stream, tranche, len(announcement.WorkReports))
	if err != nil {
		return fmt.Errorf("failed to read evidence: %w", err)
	}

	finBuf := make([]byte, 3)
	if _, err := io.ReadFull(stream, finBuf); err != nil {
		return fmt.Errorf("failed to read FIN: %w", err)
	}
	if string(finBuf) != "FIN" {
		return errors.New("request does not end with FIN")
	}

	if err := validateAuditAnnouncement(headerHash, tranche, announcement, evidence); err != nil {
		return fmt.Errorf("invalid audit announcement: %w", err)
	}
	if err := storeAuditAnnouncement(headerHash, tranche, announcement, evidence); err != nil {
		return fmt.Errorf("failed to store audit announcement: %w", err)
	}

	if _, err := stream.Write([]byte("FIN")); err != nil {
		return fmt.Errorf("failed to write FIN response: %w", err)
	}

	return stream.Close()
}

// readAnnouncement reads the announcement part of the message
func readAnnouncement(stream io.ReadWriteCloser) (*CE144Announcement, error) {
	lengthBuf := make([]byte, 4)
	if _, err := io.ReadFull(stream, lengthBuf); err != nil {
		return nil, fmt.Errorf("failed to read work reports length: %w", err)
	}
	workReportsLength := binary.LittleEndian.Uint32(lengthBuf)

	// Read work reports (Core Index + Work-Report Hash pairs)
	workReports := make([]WorkReportEntry, workReportsLength)
	for i := range workReportsLength {
		coreIndexBuf := make([]byte, 2)
		if _, err := io.ReadFull(stream, coreIndexBuf); err != nil {
			return nil, fmt.Errorf("failed to read core index %d: %w", i, err)
		}
		coreIndex := types.CoreIndex(binary.LittleEndian.Uint16(coreIndexBuf))

		workReportHash := types.WorkReportHash{}
		if _, err := io.ReadFull(stream, workReportHash[:]); err != nil {
			return nil, fmt.Errorf("failed to read work report hash %d: %w", i, err)
		}

		workReports[i] = WorkReportEntry{
			CoreIndex:      coreIndex,
			WorkReportHash: workReportHash,
		}
	}

	signature := types.Ed25519Signature{}
	if _, err := io.ReadFull(stream, signature[:]); err != nil {
		return nil, fmt.Errorf("failed to read Ed25519 signature: %w", err)
	}

	return &CE144Announcement{
		WorkReports: workReports,
		Signature:   signature,
	}, nil
}

// readEvidence reads the evidence part based on tranche
func readEvidence(stream io.ReadWriteCloser, tranche uint8, workReportsCount int) (*CE144Evidence, error) {
	if tranche == 0 {
		// First Tranche Evidence = Bandersnatch Signature (96 bytes)
		bandersnatchSig := types.BandersnatchVrfSignature{}
		if _, err := io.ReadFull(stream, bandersnatchSig[:]); err != nil {
			return nil, fmt.Errorf("failed to read Bandersnatch signature: %w", err)
		}

		return &CE144Evidence{
			IsFirstTranche:     true,
			BandersnatchSig:    bandersnatchSig,
			SubsequentEvidence: nil,
		}, nil
	} else {
		// Subsequent Tranche Evidence = [Bandersnatch Signature ++ len++[No-Show]] per work-report
		subsequentEvidence := make([]SubsequentTrancheEvidence, workReportsCount)

		for i := 0; i < workReportsCount; i++ {
			bandersnatchSig := types.BandersnatchVrfSignature{}
			if _, err := io.ReadFull(stream, bandersnatchSig[:]); err != nil {
				return nil, fmt.Errorf("failed to read Bandersnatch signature for work-report %d: %w", i, err)
			}

			// Read no-shows length
			lengthBuf := make([]byte, 4)
			if _, err := io.ReadFull(stream, lengthBuf); err != nil {
				return nil, fmt.Errorf("failed to read no-shows length for work-report %d: %w", i, err)
			}
			noShowsLength := binary.LittleEndian.Uint32(lengthBuf)

			// Read no-shows
			noShows := make([]NoShow, noShowsLength)
			for j := uint32(0); j < noShowsLength; j++ {
				// Validator Index (2 bytes)
				validatorIndexBuf := make([]byte, 2)
				if _, err := io.ReadFull(stream, validatorIndexBuf); err != nil {
					return nil, fmt.Errorf("failed to read validator index for no-show %d of work-report %d: %w", j, i, err)
				}
				validatorIndex := types.ValidatorIndex(binary.LittleEndian.Uint16(validatorIndexBuf))

				// Previous Announcement - read its length first
				prevAnnouncementLengthBuf := make([]byte, 4)
				if _, err := io.ReadFull(stream, prevAnnouncementLengthBuf); err != nil {
					return nil, fmt.Errorf("failed to read previous announcement length for no-show %d of work-report %d: %w", j, i, err)
				}
				prevAnnouncementLength := binary.LittleEndian.Uint32(prevAnnouncementLengthBuf)

				// Read previous announcement data
				prevAnnouncementData := make([]byte, prevAnnouncementLength)
				if _, err := io.ReadFull(stream, prevAnnouncementData); err != nil {
					return nil, fmt.Errorf("failed to read previous announcement data for no-show %d of work-report %d: %w", j, i, err)
				}

				noShows[j] = NoShow{
					ValidatorIndex:       validatorIndex,
					PreviousAnnouncement: prevAnnouncementData,
				}
			}

			subsequentEvidence[i] = SubsequentTrancheEvidence{
				BandersnatchSig: bandersnatchSig,
				NoShows:         noShows,
			}
		}

		return &CE144Evidence{
			IsFirstTranche:     false,
			BandersnatchSig:    types.BandersnatchVrfSignature{},
			SubsequentEvidence: subsequentEvidence,
		}, nil
	}
}

func validateAuditAnnouncement(headerHash types.OpaqueHash, tranche uint8, announcement *CE144Announcement, evidence *CE144Evidence) error {
	if len(announcement.WorkReports) == 0 {
		return errors.New("announcement must contain at least one work report")
	}

	if tranche == 0 && !evidence.IsFirstTranche {
		return errors.New("tranche 0 must have first tranche evidence")
	}
	if tranche > 0 && evidence.IsFirstTranche {
		return errors.New("non-zero tranche must have subsequent tranche evidence")
	}

	if !evidence.IsFirstTranche {
		if len(evidence.SubsequentEvidence) != len(announcement.WorkReports) {
			return fmt.Errorf("subsequent evidence count (%d) must match work reports count (%d)",
				len(evidence.SubsequentEvidence), len(announcement.WorkReports))
		}
	}

	return nil
}

func storeAuditAnnouncement(headerHash types.OpaqueHash, tranche uint8, announcement *CE144Announcement, evidence *CE144Evidence) error {
	redisBackend, err := store.GetRedisBackend()
	if err != nil {
		return fmt.Errorf("failed to get Redis backend: %w", err)
	}

	announcementData := &CE144Payload{
		HeaderHash:   headerHash,
		Tranche:      tranche,
		Announcement: *announcement,
		Evidence:     *evidence,
	}

	encodedAnnouncement, err := announcementData.Encode()
	if err != nil {
		return fmt.Errorf("failed to encode announcement data: %w", err)
	}

	headerHashHex := hex.EncodeToString(headerHash[:])
	key := fmt.Sprintf("audit_announcement:%s:%d", headerHashHex, tranche)

	client := redisBackend.GetClient()
	err = client.Put(key, encodedAnnouncement)
	if err != nil {
		return fmt.Errorf("failed to store announcement in Redis: %w", err)
	}

	headerKey := fmt.Sprintf("header_audit_announcements:%s", headerHashHex)
	err = client.SAdd(headerKey, encodedAnnouncement)
	if err != nil {
		return fmt.Errorf("failed to add announcement to header set: %w", err)
	}

	return nil
}

func CreateAuditAnnouncement(
	headerHash types.OpaqueHash,
	tranche uint8,
	announcement *CE144Announcement,
	evidence *CE144Evidence,
) ([]byte, error) {
	payload := &CE144Payload{
		HeaderHash:   headerHash,
		Tranche:      tranche,
		Announcement: *announcement,
		Evidence:     *evidence,
	}

	return payload.Encode()
}

func GetAuditAnnouncement(headerHash types.OpaqueHash, tranche uint8) (*CE144Payload, error) {
	redisBackend, err := store.GetRedisBackend()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis backend: %w", err)
	}

	headerHashHex := hex.EncodeToString(headerHash[:])
	key := fmt.Sprintf("audit_announcement:%s:%d", headerHashHex, tranche)

	client := redisBackend.GetClient()
	encodedAnnouncement, err := client.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get announcement from Redis: %w", err)
	}

	if encodedAnnouncement == nil {
		return nil, fmt.Errorf("audit announcement not found for header: %x, tranche: %d", headerHash, tranche)
	}

	announcementData := &CE144Payload{}
	err = announcementData.Decode(encodedAnnouncement)
	if err != nil {
		return nil, fmt.Errorf("failed to decode announcement data: %w", err)
	}

	return announcementData, nil
}

func GetAllAuditAnnouncementsForHeader(headerHash types.OpaqueHash) ([]*CE144Payload, error) {
	redisBackend, err := store.GetRedisBackend()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis backend: %w", err)
	}

	headerHashHex := hex.EncodeToString(headerHash[:])
	headerKey := fmt.Sprintf("header_audit_announcements:%s", headerHashHex)

	client := redisBackend.GetClient()
	encodedAnnouncements, err := client.SMembers(headerKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get announcements set from Redis: %w", err)
	}

	var announcements []*CE144Payload
	for _, encodedAnnouncement := range encodedAnnouncements {
		announcementData := &CE144Payload{}
		err := announcementData.Decode(encodedAnnouncement)
		if err != nil {
			return nil, fmt.Errorf("failed to decode announcement data: %w", err)
		}
		announcements = append(announcements, announcementData)
	}

	return announcements, nil
}

func DeleteAuditAnnouncement(headerHash types.OpaqueHash, tranche uint8) error {
	redisBackend, err := store.GetRedisBackend()
	if err != nil {
		return fmt.Errorf("failed to get Redis backend: %w", err)
	}

	headerHashHex := hex.EncodeToString(headerHash[:])
	key := fmt.Sprintf("audit_announcement:%s:%d", headerHashHex, tranche)

	client := redisBackend.GetClient()
	encodedAnnouncement, err := client.Get(key)
	if err != nil {
		return fmt.Errorf("failed to get announcement from Redis: %w", err)
	}

	if encodedAnnouncement != nil {
		headerKey := fmt.Sprintf("header_audit_announcements:%s", headerHashHex)
		err = client.SRem(headerKey, encodedAnnouncement)
		if err != nil {
			return fmt.Errorf("failed to remove announcement from header set: %w", err)
		}
	}

	err = client.Delete(key)
	if err != nil {
		return fmt.Errorf("failed to delete announcement from Redis: %w", err)
	}

	return nil
}

// Data structures for CE144

type WorkReportEntry struct {
	CoreIndex      types.CoreIndex
	WorkReportHash types.WorkReportHash
}

type CE144Announcement struct {
	WorkReports []WorkReportEntry
	Signature   types.Ed25519Signature
}

type NoShow struct {
	ValidatorIndex       types.ValidatorIndex
	PreviousAnnouncement []byte
}

type SubsequentTrancheEvidence struct {
	BandersnatchSig types.BandersnatchVrfSignature
	NoShows         []NoShow
}

type CE144Evidence struct {
	IsFirstTranche     bool
	BandersnatchSig    types.BandersnatchVrfSignature // Used for first tranche
	SubsequentEvidence []SubsequentTrancheEvidence    // Used for subsequent tranches
}

type CE144Payload struct {
	HeaderHash   types.OpaqueHash
	Tranche      uint8
	Announcement CE144Announcement
	Evidence     CE144Evidence
}

func (p *CE144Payload) Validate() error {
	if len(p.Announcement.WorkReports) == 0 {
		return errors.New("announcement must contain at least one work report")
	}

	// Validate that evidence matches the tranche type
	if p.Tranche == 0 && !p.Evidence.IsFirstTranche {
		return errors.New("tranche 0 must have first tranche evidence")
	}
	if p.Tranche > 0 && p.Evidence.IsFirstTranche {
		return errors.New("non-zero tranche must have subsequent tranche evidence")
	}

	// For subsequent tranches, validate evidence count matches work reports count
	if !p.Evidence.IsFirstTranche {
		if len(p.Evidence.SubsequentEvidence) != len(p.Announcement.WorkReports) {
			return fmt.Errorf("subsequent evidence count (%d) must match work reports count (%d)",
				len(p.Evidence.SubsequentEvidence), len(p.Announcement.WorkReports))
		}
	}

	return nil
}

func (p *CE144Payload) Encode() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	var encoded []byte

	// Header Hash (32 bytes)
	encoded = append(encoded, p.HeaderHash[:]...)

	// Tranche (1 byte)
	encoded = append(encoded, p.Tranche)

	// Announcement length (4 bytes)
	workReportsLength := make([]byte, 4)
	binary.LittleEndian.PutUint32(workReportsLength, uint32(len(p.Announcement.WorkReports)))
	encoded = append(encoded, workReportsLength...)

	// Work Reports
	for _, wr := range p.Announcement.WorkReports {
		coreIndexBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(coreIndexBytes, uint16(wr.CoreIndex))
		encoded = append(encoded, coreIndexBytes...)

		// Work Report Hash (32 bytes)
		encoded = append(encoded, wr.WorkReportHash[:]...)
	}

	encoded = append(encoded, p.Announcement.Signature[:]...)

	if p.Evidence.IsFirstTranche {
		// Bandersnatch Signature (96 bytes)
		encoded = append(encoded, p.Evidence.BandersnatchSig[:]...)
	} else {
		// Subsequent evidence for each work report
		for _, evidence := range p.Evidence.SubsequentEvidence {
			encoded = append(encoded, evidence.BandersnatchSig[:]...)

			noShowsLength := make([]byte, 4)
			binary.LittleEndian.PutUint32(noShowsLength, uint32(len(evidence.NoShows)))
			encoded = append(encoded, noShowsLength...)

			for _, noShow := range evidence.NoShows {
				validatorIndexBytes := make([]byte, 2)
				binary.LittleEndian.PutUint16(validatorIndexBytes, uint16(noShow.ValidatorIndex))
				encoded = append(encoded, validatorIndexBytes...)
				prevAnnouncementLength := make([]byte, 4)
				binary.LittleEndian.PutUint32(prevAnnouncementLength, uint32(len(noShow.PreviousAnnouncement)))
				encoded = append(encoded, prevAnnouncementLength...)
				encoded = append(encoded, noShow.PreviousAnnouncement...)
			}
		}
	}

	return encoded, nil
}

func (p *CE144Payload) Decode(data []byte) error {
	if len(data) < 37 { // HeaderHash (32) + Tranche (1) + WorkReportsLength (4)
		return fmt.Errorf("invalid data size: expected at least 37 bytes, got %d", len(data))
	}

	offset := 0

	copy(p.HeaderHash[:], data[offset:offset+32])
	offset += 32

	p.Tranche = data[offset]
	offset += 1

	workReportsLength := binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	p.Announcement.WorkReports = make([]WorkReportEntry, workReportsLength)
	for i := uint32(0); i < workReportsLength; i++ {
		if offset+34 > len(data) { // CoreIndex (2) + WorkReportHash (32)
			return fmt.Errorf("insufficient data for work report %d", i)
		}

		coreIndex := binary.LittleEndian.Uint16(data[offset : offset+2])
		offset += 2

		var workReportHash types.WorkReportHash
		copy(workReportHash[:], data[offset:offset+32])
		offset += 32

		p.Announcement.WorkReports[i] = WorkReportEntry{
			CoreIndex:      types.CoreIndex(coreIndex),
			WorkReportHash: workReportHash,
		}
	}

	if offset+64 > len(data) {
		return errors.New("insufficient data for Ed25519 signature")
	}
	copy(p.Announcement.Signature[:], data[offset:offset+64])
	offset += 64

	if p.Tranche == 0 {
		if offset+96 > len(data) {
			return errors.New("insufficient data for Bandersnatch signature")
		}
		p.Evidence.IsFirstTranche = true
		copy(p.Evidence.BandersnatchSig[:], data[offset:offset+96])
		offset += 96
	} else {
		p.Evidence.IsFirstTranche = false
		p.Evidence.SubsequentEvidence = make([]SubsequentTrancheEvidence, workReportsLength)

		for i := uint32(0); i < workReportsLength; i++ {
			if offset+96 > len(data) {
				return fmt.Errorf("insufficient data for Bandersnatch signature for work report %d", i)
			}
			copy(p.Evidence.SubsequentEvidence[i].BandersnatchSig[:], data[offset:offset+96])
			offset += 96

			if offset+4 > len(data) {
				return fmt.Errorf("insufficient data for no-shows length for work report %d", i)
			}
			noShowsLength := binary.LittleEndian.Uint32(data[offset : offset+4])
			offset += 4

			p.Evidence.SubsequentEvidence[i].NoShows = make([]NoShow, noShowsLength)
			for j := uint32(0); j < noShowsLength; j++ {
				if offset+2 > len(data) {
					return fmt.Errorf("insufficient data for validator index for no-show %d of work report %d", j, i)
				}
				validatorIndex := binary.LittleEndian.Uint16(data[offset : offset+2])
				offset += 2

				if offset+4 > len(data) {
					return fmt.Errorf("insufficient data for previous announcement length for no-show %d of work report %d", j, i)
				}
				prevAnnouncementLength := binary.LittleEndian.Uint32(data[offset : offset+4])
				offset += 4

				if offset+int(prevAnnouncementLength) > len(data) {
					return fmt.Errorf("insufficient data for previous announcement data for no-show %d of work report %d", j, i)
				}
				prevAnnouncementData := make([]byte, prevAnnouncementLength)
				copy(prevAnnouncementData, data[offset:offset+int(prevAnnouncementLength)])
				offset += int(prevAnnouncementLength)

				p.Evidence.SubsequentEvidence[i].NoShows[j] = NoShow{
					ValidatorIndex:       types.ValidatorIndex(validatorIndex),
					PreviousAnnouncement: prevAnnouncementData,
				}
			}
		}
	}

	return p.Validate()
}
