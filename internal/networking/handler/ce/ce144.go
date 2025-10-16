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
	vrf "github.com/New-JAMneration/JAM-Protocol/pkg/Rust-VRF/vrf-func-ffi/src"
)

// HandleAuditAnnouncement_Recv handles the announcement of requirement to audit.
func HandleAuditAnnouncement_Recv(blockchain blockchain.Blockchain, stream *quic.Stream) error {
	// First message: Header Hash ++ Tranche ++ Announcement
	firstMessage := make([]byte, 33) // 32 bytes header hash + 1 byte tranche
	if err := stream.ReadFull(firstMessage); err != nil {
		return fmt.Errorf("failed to read header hash and tranche: %w", err)
	}

	var headerHash types.HeaderHash
	copy(headerHash[:], firstMessage[:32])
	tranche := uint8(firstMessage[32])

	announcement, err := readAnnouncement(stream)
	if err != nil {
		return fmt.Errorf("failed to read announcement: %w", err)
	}

	// Second message: Evidence
	evidence, err := readEvidence(stream, tranche, len(announcement.WorkReports))
	if err != nil {
		return fmt.Errorf("failed to read evidence: %w", err)
	}

	// If this is the first tranche, verify the Bandersnatch IETF-VRF signature (96 bytes)
	if tranche == 0 {
		// Build ring from posterior validators
		s := store.GetInstance()
		gamma := s.GetPosteriorStates().GetGammaK()
		if len(gamma) == 0 {
			return fmt.Errorf("empty gamma for verifier")
		}

		ring := []byte{}
		for _, v := range gamma {
			ring = append(ring, v.Bandersnatch[:]...)
		}

		verifier, err := vrf.NewVerifier(ring, uint(len(gamma)))
		if err != nil {
			return fmt.Errorf("failed to create vrf verifier: %w", err)
		}
		defer verifier.Free()

		// Serialize announcement into bytes to use as message (must match signer semantics)
		tmpPayload := &CE144Payload{
			HeaderHash:   headerHash,
			Tranche:      0,
			Announcement: *announcement,
			Evidence:     CE144Evidence{},
		}
		annBytes, err := tmpPayload.Encode()
		if err != nil {
			return fmt.Errorf("failed to encode announcement for verification: %w", err)
		}

		ctx := []byte{}
		msg := annBytes

		verified := false
		for i := uint(0); i < uint(len(gamma)); i++ {
			if _, verr := verifier.IETFVerify(ctx, msg, evidence.BandersnatchSig[:], i); verr == nil {
				verified = true
				break
			}
		}
		if !verified {
			return fmt.Errorf("bandersnatch signature verification failed for announcement")
		}
	}

	// Third message: FIN
	finBuf := make([]byte, 3)
	if err := stream.ReadFull(finBuf); err != nil {
		return fmt.Errorf("failed to read FIN: %w", err)
	} else if string(finBuf) != "FIN" {
		return errors.New("request does not end with FIN")
	}

	if err := validateAuditAnnouncement(headerHash, tranche, announcement, evidence); err != nil {
		return fmt.Errorf("invalid audit announcement: %w", err)
	}
	if err := storeAuditAnnouncement(headerHash, tranche, announcement, evidence); err != nil {
		return fmt.Errorf("failed to store audit announcement: %w", err)
	}

	if err := stream.WriteMessage([]byte("FIN")); err != nil {
		return fmt.Errorf("failed to write FIN response: %w", err)
	}

	return stream.Close()
}

// readAnnouncement reads the announcement part of the message
func readAnnouncement(stream *quic.Stream) (*CE144Announcement, error) {
	lengthBuf := make([]byte, 4)
	if err := stream.ReadFull(lengthBuf); err != nil {
		return nil, fmt.Errorf("failed to read work reports length: %w", err)
	}
	workReportsLength := binary.LittleEndian.Uint32(lengthBuf)

	// Read work reports (Core Index + Work-Report Hash pairs)
	workReports := make([]WorkReportEntry, workReportsLength)
	for i := uint32(0); i < workReportsLength; i++ {
		coreIndexBuf := make([]byte, 2)
		if err := stream.ReadFull(coreIndexBuf); err != nil {
			return nil, fmt.Errorf("failed to read core index %d: %w", i, err)
		}
		coreIndex := types.CoreIndex(binary.LittleEndian.Uint16(coreIndexBuf))

		workReportHash := types.WorkReportHash{}
		if err := stream.ReadFull(workReportHash[:]); err != nil {
			return nil, fmt.Errorf("failed to read work report hash %d: %w", i, err)
		}

		workReports[i] = WorkReportEntry{
			CoreIndex:      coreIndex,
			WorkReportHash: workReportHash,
		}
	}

	signature := types.Ed25519Signature{}
	if _, err := stream.Read(signature[:]); err != nil {
		return nil, fmt.Errorf("failed to read Ed25519 signature: %w", err)
	}

	return &CE144Announcement{
		WorkReports: workReports,
		Signature:   signature,
	}, nil
}

// readEvidence reads the evidence part based on tranche
func readEvidence(stream *quic.Stream, tranche uint8, workReportsCount int) (*CE144Evidence, error) {
	if tranche == 0 {
		// First Tranche Evidence = Bandersnatch Signature (96 bytes)
		bandersnatchSig := types.BandersnatchVrfSignature{}
		if _, err := stream.Read(bandersnatchSig[:]); err != nil {
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
			if err := stream.ReadFull(bandersnatchSig[:]); err != nil {
				return nil, fmt.Errorf("failed to read Bandersnatch signature for work-report %d: %w", i, err)
			}

			// Read no-shows length
			lengthBuf := make([]byte, 4)
			if err := stream.ReadFull(lengthBuf); err != nil {
				return nil, fmt.Errorf("failed to read no-shows length for work-report %d: %w", i, err)
			}
			noShowsLength := binary.LittleEndian.Uint32(lengthBuf)

			// Read no-shows
			noShows := make([]NoShow, noShowsLength)
			for j := uint32(0); j < noShowsLength; j++ {
				// Validator Index (2 bytes)
				validatorIndexBuf := make([]byte, 2)
				if err := stream.ReadFull(validatorIndexBuf); err != nil {
					return nil, fmt.Errorf("failed to read validator index for no-show %d of work-report %d: %w", j, i, err)
				}
				validatorIndex := types.ValidatorIndex(binary.LittleEndian.Uint16(validatorIndexBuf))

				// Previous Announcement - read its length first
				prevAnnouncementLengthBuf := make([]byte, 4)
				if err := stream.ReadFull(prevAnnouncementLengthBuf); err != nil {
					return nil, fmt.Errorf("failed to read previous announcement length for no-show %d of work-report %d: %w", j, i, err)
				}
				prevAnnouncementLength := binary.LittleEndian.Uint32(prevAnnouncementLengthBuf)

				// Read previous announcement data
				prevAnnouncementData := make([]byte, prevAnnouncementLength)
				if err := stream.ReadFull(prevAnnouncementData); err != nil {
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

func validateAuditAnnouncement(headerHash types.HeaderHash, tranche uint8, announcement *CE144Announcement, evidence *CE144Evidence) error {
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

func storeAuditAnnouncement(headerHash types.HeaderHash, tranche uint8, announcement *CE144Announcement, evidence *CE144Evidence) error {
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

// HandleAuditAnnouncement_Send sends an audit announcement
func HandleAuditAnnouncement_Send(stream *quic.Stream, headerHash types.HeaderHash, tranche uint8, announcement *CE144Announcement, evidence *CE144Evidence) error {

	if err := stream.WriteMessage([]byte{144}); err != nil {
		return fmt.Errorf("failed to write protocol ID: %w", err)
	}

	annBytes, err := encodeAnnouncementPart(headerHash, tranche, announcement)
	if err != nil {
		return fmt.Errorf("failed to encode announcement part: %w", err)
	}
	if err := stream.WriteMessage(annBytes); err != nil {
		return fmt.Errorf("failed to write announcement part: %w", err)
	}

	evidenceBytes, err := encodeEvidencePart(tranche, evidence, len(announcement.WorkReports))
	if err != nil {
		return fmt.Errorf("failed to encode evidence part: %w", err)
	}
	if err := stream.WriteMessage(evidenceBytes); err != nil {
		return fmt.Errorf("failed to write evidence part: %w", err)
	}

	if err := stream.WriteMessage([]byte("FIN")); err != nil {
		return fmt.Errorf("failed to write FIN: %w", err)
	}

	finBuf := make([]byte, 3)
	n, err := stream.Read(finBuf)
	if err != nil {
		return fmt.Errorf("failed to read FIN response: %w", err)
	}
	if n != 3 || string(finBuf[:n]) != "FIN" {
		return fmt.Errorf("unexpected FIN response: %q", string(finBuf[:n]))
	}

	return stream.Close()
}

// encodeAnnouncementPart serializes the HeaderHash + Tranche + Announcement (without evidence)
func encodeAnnouncementPart(headerHash types.HeaderHash, tranche uint8, announcement *CE144Announcement) ([]byte, error) {
	if announcement == nil {
		return nil, fmt.Errorf("nil announcement")
	}

	var encoded []byte
	// Header Hash (32 bytes)
	encoded = append(encoded, headerHash[:]...)
	// Tranche (1 byte)
	encoded = append(encoded, tranche)

	// Announcement length (4 bytes)
	workReportsLength := make([]byte, 4)
	binary.LittleEndian.PutUint32(workReportsLength, uint32(len(announcement.WorkReports)))
	encoded = append(encoded, workReportsLength...)

	// Work Reports
	for _, wr := range announcement.WorkReports {
		coreIndexBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(coreIndexBytes, uint16(wr.CoreIndex))
		encoded = append(encoded, coreIndexBytes...)
		encoded = append(encoded, wr.WorkReportHash[:]...)
	}

	// Ed25519 signature
	encoded = append(encoded, announcement.Signature[:]...)

	return encoded, nil
}

// encodeEvidencePart serializes the evidence part depending on tranche
func encodeEvidencePart(tranche uint8, evidence *CE144Evidence, workReportsCount int) ([]byte, error) {
	if evidence == nil {
		return nil, fmt.Errorf("nil evidence")
	}

	var encoded []byte
	if tranche == 0 {
		// First tranche: just Bandersnatch signature (96 bytes)
		encoded = append(encoded, evidence.BandersnatchSig[:]...)
		return encoded, nil
	}

	// Subsequent tranche: for each work report emit BandersnatchSig ++ len(noShows) ++ noShows
	if len(evidence.SubsequentEvidence) != workReportsCount {
		return nil, fmt.Errorf("subsequent evidence count (%d) does not match work reports count (%d)", len(evidence.SubsequentEvidence), workReportsCount)
	}

	for _, se := range evidence.SubsequentEvidence {
		encoded = append(encoded, se.BandersnatchSig[:]...)

		noShowsLength := make([]byte, 4)
		binary.LittleEndian.PutUint32(noShowsLength, uint32(len(se.NoShows)))
		encoded = append(encoded, noShowsLength...)

		for _, ns := range se.NoShows {
			validatorIndexBytes := make([]byte, 2)
			binary.LittleEndian.PutUint16(validatorIndexBytes, uint16(ns.ValidatorIndex))
			encoded = append(encoded, validatorIndexBytes...)

			prevAnnouncementLength := make([]byte, 4)
			binary.LittleEndian.PutUint32(prevAnnouncementLength, uint32(len(ns.PreviousAnnouncement)))
			encoded = append(encoded, prevAnnouncementLength...)
			encoded = append(encoded, ns.PreviousAnnouncement...)
		}
	}

	return encoded, nil
}

// test helper functions

func CreateAuditAnnouncement(
	headerHash types.HeaderHash,
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

func GetAuditAnnouncement(headerHash types.HeaderHash, tranche uint8) (*CE144Payload, error) {
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

func GetAllAuditAnnouncementsForHeader(headerHash types.HeaderHash) ([]*CE144Payload, error) {
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

func (h *DefaultCERequestHandler) encodeAuditAnnouncement(message interface{}) ([]byte, error) {
	announcement, ok := message.(*CE144Payload)
	if !ok {
		return nil, fmt.Errorf("unsupported message type for AuditAnnouncement: %T", message)
	}

	if announcement == nil {
		return nil, fmt.Errorf("nil payload for AuditAnnouncement")
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
	HeaderHash   types.HeaderHash
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
