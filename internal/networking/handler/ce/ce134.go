package ce

import (
	"crypto/ed25519"
	"encoding/binary"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/store/keystore"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/New-JAMneration/JAM-Protocol/internal/work_package"
)

type CE134Payload struct {
	CoreIndex   uint16
	HeaderHash  [32]byte
	WorkPackage *types.WorkPackage
}

type SegmentRootMapping struct {
	WorkPackageHash types.WorkPackageHash
	SegmentRoot     types.OpaqueHash
}

// Helper to decode segment-root mappings from the stream
func readSegmentRootMappings(stream *quic.Stream) ([]SegmentRootMapping, error) {
	var mappings []SegmentRootMapping
	var countBuf [1]byte
	if err := stream.ReadFull(countBuf[:]); err != nil {
		return nil, err
	}
	count := int(countBuf[0])
	for range count {
		var wpHash types.WorkPackageHash
		var segRoot types.OpaqueHash
		if err := stream.ReadFull(wpHash[:]); err != nil {
			return nil, err
		}
		if err := stream.ReadFull(segRoot[:]); err != nil {
			return nil, err
		}
		mappings = append(mappings, SegmentRootMapping{wpHash, segRoot})
	}
	return mappings, nil
}

// Role: [Guarantor -> Guarantor]
//
// [TODO-Validation]
// 1. Verify work-report.
// 2. Ensure all import segments retrieved.
// 3. Ensure work-package bundle exactly matches erasure-coded bundle.
// 4. Add basic verification (auth + mappings) after receiving bundle.
func HandleWorkPackageShare_Recv(
	blockchain blockchain.Blockchain,
	stream *quic.Stream,
	keypair keystore.KeyPair,
	pvmExecutor work_package.PVMExecutor,
	erasureMap *store.SegmentErasureMap,
	segmentRootLookup *store.HashSegmentMap,
) error {
	// 1. Read core index (2 bytes, little endian)
	coreIndexBuf := make([]byte, 2)
	if _, err := stream.Read(coreIndexBuf); err != nil {
		return err
	}
	coreIndex := types.CoreIndex(binary.LittleEndian.Uint16(coreIndexBuf))

	// 2. Read segment-root mappings
	_, err := readSegmentRootMappings(stream)
	if err != nil {
		return fmt.Errorf("failed to read segment-root mappings: %w", err)
	}

	// 3. Read work-package bundle (rest of stream until FIN)
	bundle := make([]byte, 65536)

	if err := stream.ReadFull(bundle); err != nil {
		return fmt.Errorf("failed to read bundle: %w", err)
	}

	controller := work_package.NewSharedController(bundle, erasureMap, segmentRootLookup, coreIndex)
	if pvmExecutor != nil {
		controller.PVM = pvmExecutor
	}
	workReport, err := controller.Process()
	if err != nil {
		return fmt.Errorf("work-package failed to process into work-report: %w", err)
	}

	workReportSerialization := utilities.WorkReportSerialization(workReport)
	workReportHash := hash.Blake2bHash(workReportSerialization)

	// message: JamGuarantee context + work report hash
	message := []byte(types.JamGuarantee)
	message = append(message, workReportHash[:]...)

	sig := ed25519.Sign(keypair.PrivateKey(), message)
	resp := append(workReportHash[:], sig...)
	if _, err := stream.Write(resp); err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}
	return stream.Close()
}

// HandleWorkPackageShare_Send implements the sender side of Guarantor -> Guarantor sharing.
func HandleWorkPackageShare_Send(
	stream *quic.Stream,
	coreIndex types.CoreIndex,
	mappings []SegmentRootMapping,
	bundle []byte,
	peerEd25519 types.Ed25519Public,
) (types.WorkReportHash, []byte, error) {
	// Build and write: Core Index (LE16) ++ Mappings (len:u8 ++ [WP Hash ++ Segments-Root]) in one write
	if len(mappings) > 255 {
		var zero types.WorkReportHash
		return zero, nil, fmt.Errorf("too many mappings: %d", len(mappings))
	}
	headerAndMappings := make([]byte, 0, 2+1+len(mappings)*64)
	headerAndMappings = append(headerAndMappings, encodeLE16(uint16(coreIndex))...)
	headerAndMappings = append(headerAndMappings, byte(len(mappings)))
	for _, m := range mappings {
		headerAndMappings = append(headerAndMappings, m.WorkPackageHash[:]...)
		headerAndMappings = append(headerAndMappings, m.SegmentRoot[:]...)
	}
	if err := stream.WriteMessage(headerAndMappings); err != nil {
		var zero types.WorkReportHash
		return zero, nil, fmt.Errorf("failed to write header and mappings: %w", err)
	}

	if len(bundle) > 0 {
		if err := stream.WriteMessage(bundle); err != nil {
			return types.WorkReportHash{}, nil, fmt.Errorf("failed to write bundle: %w", err)
		}
	}

	if err := stream.WriteMessage([]byte("FIN")); err != nil {
		return types.WorkReportHash{}, nil, fmt.Errorf("failed to write FIN: %w", err)
	}

	// Read response: Work-Report Hash (32) + Signature (64)
	resp := make([]byte, 32+ed25519.SignatureSize)
	if err := stream.ReadFull(resp); err != nil {
		return types.WorkReportHash{}, nil, fmt.Errorf("failed to read response: %w", err)
	}
	var wrHash types.WorkReportHash
	copy(wrHash[:], resp[:32])
	sig := make([]byte, ed25519.SignatureSize)
	copy(sig, resp[32:])

	msg := []byte(types.JamGuarantee)
	msg = append(msg, wrHash[:]...)
	if !ed25519.Verify(ed25519.PublicKey(peerEd25519[:]), msg, sig) {
		return types.WorkReportHash{}, nil, fmt.Errorf("bad_guarantor_signature")
	}

	return wrHash, sig, stream.Close()
}

func (h *DefaultCERequestHandler) encodeWorkPackageSharing(message interface{}) ([]byte, error) {
	workPackage, ok := message.(*CE134Payload)
	if !ok {
		return nil, fmt.Errorf("unsupported message type for WorkPackageSharing: %T", message)
	}

	if workPackage == nil {
		return nil, fmt.Errorf("nil payload for WorkPackageSharing")
	}

	if workPackage.WorkPackage == nil {
		return nil, fmt.Errorf("nil WorkPackage in CE134Payload")
	}

	requestType := byte(WorkPackageSharing)

	coreIndexBytes := encodeLE16(workPackage.CoreIndex)

	// Get WorkPackage bytes using ScaleEncode
	wpBytes, err := workPackage.WorkPackage.ScaleEncode()
	if err != nil {
		return nil, fmt.Errorf("failed to encode WorkPackage: %w", err)
	}

	totalLen := 1 + 4 + 32 + len(wpBytes) // 1 byte for request type + 4 bytes for CoreIndex + 32 bytes for HeaderHash + WorkPackage bytes

	result := make([]byte, 0, totalLen)

	result = append(result, requestType)
	result = append(result, coreIndexBytes...)
	result = append(result, workPackage.HeaderHash[:]...)
	result = append(result, wpBytes...)

	return result, nil
}
