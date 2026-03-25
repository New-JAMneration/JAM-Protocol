package ce

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// ce144Stream supports JAMNP message framing (ReadMessage / WriteMessage).
type ce144Stream interface {
	io.ReadWriteCloser
	ReadMessage() ([]byte, error)
	WriteMessage(payload []byte) error
}

// HandleAuditAnnouncement_Send sends CE144 (AuditAnnouncement) over a stream.
//
// Sends two JAMNP-framed messages then FIN:
//
//	Msg 1 → Header Hash ++ Tranche ++ len++[Core Index ++ Work-Report Hash] ++ Ed25519 Signature
//	Msg 2 → Evidence
//
// NOTE: Callers must open the stream and set the correct stream kind.
func HandleAuditAnnouncement_Send(stream ce144Stream, payload *CE144Payload) error {
	if payload == nil {
		return fmt.Errorf("nil payload")
	}
	if err := payload.Validate(); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}

	if _, err := stream.Write([]byte{144}); err != nil {
		return fmt.Errorf("failed to write protocol ID: %w", err)
	}

	msg1, err := payload.encodeMsg1()
	if err != nil {
		return fmt.Errorf("failed to encode announcement: %w", err)
	}
	if err := stream.WriteMessage(msg1); err != nil {
		return fmt.Errorf("failed to write announcement: %w", err)
	}

	msg2, err := payload.encodeMsg2()
	if err != nil {
		return fmt.Errorf("failed to encode evidence: %w", err)
	}
	if err := stream.WriteMessage(msg2); err != nil {
		return fmt.Errorf("failed to write evidence: %w", err)
	}

	if err := expectRemoteFIN(stream); err != nil {
		return err
	}
	return stream.Close()
}

// HandleAuditAnnouncement handles the announcement of requirement to audit.
// Auditors of a block (defined to be the prior validator set) should, at the beginning
// of each tranche, broadcast an announcement to all other such auditors specifying
// which work-reports they intend to audit.
//
// Protocol CE144:
// Auditor -> Auditor
//
//	--> Header Hash ++ Tranche ++ Announcement  (msg 1)
//	--> Evidence                                (msg 2)
//	--> FIN
//	<-- FIN
//
// Announcement = len++[Core Index ++ Work-Report Hash] ++ Ed25519 Signature
// Evidence     = Bandersnatch Sig (tranche 0) OR
//
//	per-work-report[Bandersnatch Sig ++ len++[No-Show]] (tranche > 0)
func HandleAuditAnnouncement(blockchain blockchain.Blockchain, stream ce144Stream) error {
	msg1, err := stream.ReadMessage()
	if err != nil {
		return fmt.Errorf("failed to read announcement message: %w", err)
	}
	headerHash, tranche, announcement, _, err := parseMsg1(msg1)
	if err != nil {
		return fmt.Errorf("failed to parse announcement: %w", err)
	}

	msg2, err := stream.ReadMessage()
	if err != nil {
		return fmt.Errorf("failed to read evidence message: %w", err)
	}
	evidence, err := parseMsg2(msg2, tranche, len(announcement.WorkReports))
	if err != nil {
		return fmt.Errorf("failed to parse evidence: %w", err)
	}

	if err := expectRemoteFIN(stream); err != nil {
		return err
	}

	if err := validateAuditAnnouncement(headerHash, tranche, announcement, evidence); err != nil {
		return fmt.Errorf("invalid audit announcement: %w", err)
	}
	if err := storeAuditAnnouncement(blockchain, headerHash, tranche, announcement, evidence); err != nil {
		return fmt.Errorf("failed to store audit announcement: %w", err)
	}
	return stream.Close()
}

// ── Wire-format codec helpers ──────────────────────────────────────────────────

// encodeMsg1 encodes CE144 message 1 bytes.
// Wire: HeaderHash(32) ++ Tranche(1) ++ len++[CoreIndex(2) ++ WorkReportHash(32)] ++ Ed25519Sig(64)
func (p *CE144Payload) encodeMsg1() ([]byte, error) {
	var buf []byte
	buf = append(buf, p.HeaderHash[:]...)
	buf = append(buf, p.Tranche)

	countBytes, err := types.NewEncoder().EncodeUint(uint64(len(p.Announcement.WorkReports)))
	if err != nil {
		return nil, fmt.Errorf("failed to encode work reports count: %w", err)
	}
	buf = append(buf, countBytes...)

	for _, wr := range p.Announcement.WorkReports {
		buf = binary.LittleEndian.AppendUint16(buf, uint16(wr.CoreIndex))
		buf = append(buf, wr.WorkReportHash[:]...)
	}
	buf = append(buf, p.Announcement.Signature[:]...)
	return buf, nil
}

// encodeMsg2 encodes CE144 message 2 bytes (evidence).
// Tranche 0:  Bandersnatch Signature (96 bytes)
// Tranche >0: per work-report → Bandersnatch Sig ++ len++[ValidatorIndex(2) ++ len++[PreviousAnnouncement]]
func (p *CE144Payload) encodeMsg2() ([]byte, error) {
	var buf []byte
	if p.Evidence.IsFirstTranche {
		buf = append(buf, p.Evidence.BandersnatchSig[:]...)
		return buf, nil
	}
	for _, ev := range p.Evidence.SubsequentEvidence {
		buf = append(buf, ev.BandersnatchSig[:]...)

		nsCountBytes, err := types.NewEncoder().EncodeUint(uint64(len(ev.NoShows)))
		if err != nil {
			return nil, fmt.Errorf("failed to encode no-shows count: %w", err)
		}
		buf = append(buf, nsCountBytes...)

		for _, ns := range ev.NoShows {
			buf = binary.LittleEndian.AppendUint16(buf, uint16(ns.ValidatorIndex))

			prevLenBytes, err := types.NewEncoder().EncodeUint(uint64(len(ns.PreviousAnnouncement)))
			if err != nil {
				return nil, fmt.Errorf("failed to encode previous announcement length: %w", err)
			}
			buf = append(buf, prevLenBytes...)
			buf = append(buf, ns.PreviousAnnouncement...)
		}
	}
	return buf, nil
}

// parseMsg1 parses CE144 message 1 bytes into its components.
// Returns (headerHash, tranche, announcement, bytesConsumed, error).
// bytesConsumed is useful for Decode() which processes msg1 + msg2 as a flat blob.
func parseMsg1(data []byte) (headerHash types.OpaqueHash, tranche uint8, announcement *CE144Announcement, consumed int, err error) {
	const minSize = HashSize + U8Size + U8Size + types.Ed25519SigSize // 32+1+1(min len++)+64 = 98
	announcement = nil
	consumed = 0

	if len(data) < minSize {
		return headerHash, tranche, announcement, consumed, fmt.Errorf("msg1 too short: %d bytes", len(data))
	}

	offset := 0
	copy(headerHash[:], data[offset:offset+HashSize])
	offset += HashSize

	tranche = data[offset]
	offset++

	count, n, err := decodeCompactUint(data[offset:])
	if err != nil {
		return headerHash, tranche, announcement, consumed, fmt.Errorf("failed to decode work reports count: %w", err)
	}
	offset += n

	workReports := make([]WorkReportEntry, count)
	for i := uint64(0); i < count; i++ {
		if offset+U16Size+HashSize > len(data) {
			return headerHash, tranche, announcement, consumed, fmt.Errorf("insufficient data for work report %d", i)
		}
		coreIndex := types.CoreIndex(binary.LittleEndian.Uint16(data[offset:]))
		offset += U16Size
		var wrHash types.WorkReportHash
		copy(wrHash[:], data[offset:offset+HashSize])
		offset += HashSize
		workReports[i] = WorkReportEntry{CoreIndex: coreIndex, WorkReportHash: wrHash}
	}

	if offset+types.Ed25519SigSize > len(data) {
		return headerHash, tranche, announcement, consumed, errors.New("insufficient data for Ed25519 signature")
	}
	var sig types.Ed25519Signature
	copy(sig[:], data[offset:offset+types.Ed25519SigSize])
	offset += types.Ed25519SigSize

	return headerHash, tranche, &CE144Announcement{WorkReports: workReports, Signature: sig}, offset, nil
}

// parseMsg2 parses CE144 message 2 (evidence) bytes.
func parseMsg2(data []byte, tranche uint8, workReportsCount int) (*CE144Evidence, error) {
	if tranche == 0 {
		if len(data) != types.BandersnatchSigSize {
			return nil, fmt.Errorf("invalid first-tranche evidence size: expected %d, got %d", types.BandersnatchSigSize, len(data))
		}
		var bsSig types.BandersnatchVrfSignature
		copy(bsSig[:], data[:types.BandersnatchSigSize])
		return &CE144Evidence{IsFirstTranche: true, BandersnatchSig: bsSig}, nil
	}

	offset := 0
	subEvidence := make([]SubsequentTrancheEvidence, workReportsCount)
	for i := 0; i < workReportsCount; i++ {
		if offset+types.BandersnatchSigSize > len(data) {
			return nil, fmt.Errorf("insufficient data for bandersnatch sig for work-report %d", i)
		}
		var bsSig types.BandersnatchVrfSignature
		copy(bsSig[:], data[offset:offset+types.BandersnatchSigSize])
		offset += types.BandersnatchSigSize

		nsCount, n, err := decodeCompactUint(data[offset:])
		if err != nil {
			return nil, fmt.Errorf("failed to decode no-shows count for work-report %d: %w", i, err)
		}
		offset += n

		noShows := make([]NoShow, nsCount)
		for j := uint64(0); j < nsCount; j++ {
			if offset+U16Size > len(data) {
				return nil, fmt.Errorf("insufficient data for validator index for no-show %d of work-report %d", j, i)
			}
			validatorIndex := types.ValidatorIndex(binary.LittleEndian.Uint16(data[offset:]))
			offset += U16Size

			prevAnnLen, n, err := decodeCompactUint(data[offset:])
			if err != nil {
				return nil, fmt.Errorf("failed to decode previous announcement length for no-show %d of work-report %d: %w", j, i, err)
			}
			offset += n

			if offset+int(prevAnnLen) > len(data) {
				return nil, fmt.Errorf("insufficient data for previous announcement for no-show %d of work-report %d", j, i)
			}
			prevAnn := make([]byte, prevAnnLen)
			copy(prevAnn, data[offset:offset+int(prevAnnLen)])
			offset += int(prevAnnLen)

			noShows[j] = NoShow{
				ValidatorIndex:       validatorIndex,
				PreviousAnnouncement: prevAnn,
			}
		}
		subEvidence[i] = SubsequentTrancheEvidence{
			BandersnatchSig: bsSig,
			NoShows:         noShows,
		}
	}
	return &CE144Evidence{IsFirstTranche: false, SubsequentEvidence: subEvidence}, nil
}

// ── Validation ────────────────────────────────────────────────────────────────

func validateAuditAnnouncement(headerHash types.OpaqueHash, tranche uint8, announcement *CE144Announcement, evidence *CE144Evidence) error {
	// We should validate headerHash by checking if this headerHash is in the database

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
		// Each no-show must carry a non-empty previous announcement blob.
		for i, ev := range evidence.SubsequentEvidence {
			for j, ns := range ev.NoShows {
				if len(ns.PreviousAnnouncement) == 0 {
					return fmt.Errorf("no-show %d of work report %d has empty previous announcement", j, i)
				}
			}
		}
	}

	return nil
}

func storeAuditAnnouncement(bc blockchain.Blockchain, headerHash types.OpaqueHash, tranche uint8, announcement *CE144Announcement, evidence *CE144Evidence) error {
	db := DB(bc)
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

	if err := PutKV(db, ceAuditAnnKey(headerHash, tranche), encodedAnnouncement); err != nil {
		return fmt.Errorf("failed to store announcement: %w", err)
	}
	if err := SAdd(db, ceAuditAnnHeaderSetKey(headerHash), encodedAnnouncement); err != nil {
		return fmt.Errorf("failed to add announcement to header set: %w", err)
	}
	return nil
}

// ── Public API ─────────────────────────────────────────────────────────────────

// CreateAuditAnnouncement builds the wire-format bytes for a CE144 stream:
// two consecutive JAMNP-framed messages (msg1 = announcement, msg2 = evidence).
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
	if err := payload.Validate(); err != nil {
		return nil, err
	}

	msg1, err := payload.encodeMsg1()
	if err != nil {
		return nil, fmt.Errorf("failed to encode announcement: %w", err)
	}
	msg2, err := payload.encodeMsg2()
	if err != nil {
		return nil, fmt.Errorf("failed to encode evidence: %w", err)
	}

	var result bytes.Buffer
	if err := quic.WriteMessageFrame(&result, msg1); err != nil {
		return nil, fmt.Errorf("failed to frame announcement: %w", err)
	}
	if err := quic.WriteMessageFrame(&result, msg2); err != nil {
		return nil, fmt.Errorf("failed to frame evidence: %w", err)
	}
	return result.Bytes(), nil
}

func GetAuditAnnouncement(bc blockchain.Blockchain, headerHash types.OpaqueHash, tranche uint8) (*CE144Payload, error) {
	db := DB(bc)
	encodedAnnouncement, err := GetKV(db, ceAuditAnnKey(headerHash, tranche))
	if err != nil {
		return nil, fmt.Errorf("failed to get announcement: %w", err)
	}
	if encodedAnnouncement == nil {
		return nil, fmt.Errorf("audit announcement not found for header: %x, tranche: %d", headerHash, tranche)
	}
	announcementData := &CE144Payload{}
	if err := announcementData.Decode(encodedAnnouncement); err != nil {
		return nil, fmt.Errorf("failed to decode announcement data: %w", err)
	}
	return announcementData, nil
}

func GetAllAuditAnnouncementsForHeader(bc blockchain.Blockchain, headerHash types.OpaqueHash) ([]*CE144Payload, error) {
	db := DB(bc)
	encodedAnnouncements, err := SMembers(db, ceAuditAnnHeaderSetKey(headerHash))
	if err != nil {
		return nil, fmt.Errorf("failed to get announcements set: %w", err)
	}
	var announcements []*CE144Payload
	for _, encodedAnnouncement := range encodedAnnouncements {
		announcementData := &CE144Payload{}
		if err := announcementData.Decode(encodedAnnouncement); err != nil {
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
	return announcement.Encode()
}

// ── Data Structures ────────────────────────────────────────────────────────────

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
		for i, ev := range p.Evidence.SubsequentEvidence {
			for j, ns := range ev.NoShows {
				if len(ns.PreviousAnnouncement) == 0 {
					return fmt.Errorf("no-show %d of work report %d has empty previous announcement", j, i)
				}
			}
		}
	}

	return nil
}

// Encode serialises the payload as a flat byte blob (for storage).
// Uses compact len++ encoding for sequence lengths, consistent with the wire format.
func (p *CE144Payload) Encode() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	msg1, err := p.encodeMsg1()
	if err != nil {
		return nil, err
	}
	msg2, err := p.encodeMsg2()
	if err != nil {
		return nil, err
	}
	return append(msg1, msg2...), nil
}

// Decode deserialises a flat byte blob (as produced by Encode) into the payload.
// Uses compact len++ decoding for sequence lengths.
func (p *CE144Payload) Decode(data []byte) error {
	headerHash, tranche, announcement, n, err := parseMsg1(data)
	if err != nil {
		return fmt.Errorf("failed to parse announcement section: %w", err)
	}
	evidence, err := parseMsg2(data[n:], tranche, len(announcement.WorkReports))
	if err != nil {
		return fmt.Errorf("failed to parse evidence section: %w", err)
	}
	p.HeaderHash = headerHash
	p.Tranche = tranche
	p.Announcement = *announcement
	p.Evidence = *evidence
	return p.Validate()
}
