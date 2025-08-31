package ce

import (
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/store/keystore"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
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
	if _, err := io.ReadFull(stream, countBuf[:]); err != nil {
		return nil, err
	}
	count := int(countBuf[0])
	for i := 0; i < count; i++ {
		var wpHash types.WorkPackageHash
		var segRoot types.OpaqueHash
		if _, err := io.ReadFull(stream, wpHash[:]); err != nil {
			return nil, err
		}
		if _, err := io.ReadFull(stream, segRoot[:]); err != nil {
			return nil, err
		}
		mappings = append(mappings, SegmentRootMapping{wpHash, segRoot})
	}
	return mappings, nil
}

// Handler for CE 134: Guarantor <-> Guarantor work-package sharing
//
// [TODO-Validation]
// 1. Verify work-report.
// 2. Ensure all import segments retrieved.
// 3. Ensure work-package bundle exactly matches erasure-coded bundle.
// 4. Add basic verification (auth + mappings) after receiving bundle.
func HandleWorkPackageShare(
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
	n, err := stream.Read(bundle)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read bundle: %w", err)
	}
	bundle = bundle[:n]

	controller := work_package.NewSharedController(bundle, erasureMap, segmentRootLookup, coreIndex)
	if pvmExecutor != nil {
		controller.PVM = pvmExecutor
	}
	workReport, err := controller.Process()
	if err != nil {
		return fmt.Errorf("work-package failed to process into work-report: %w", err)
	}

	workReportHash := workReport.PackageSpec.Hash
	message := []byte{}
	message = append(message, workReportHash[:]...)

	sig := ed25519.Sign(keypair.PrivateKey(), message)
	resp := append(workReportHash[:], sig...)
	if _, err := stream.Write(resp); err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}
	return stream.Close()
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
