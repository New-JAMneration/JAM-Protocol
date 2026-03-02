package ce

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/store/keystore"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/New-JAMneration/JAM-Protocol/internal/work_package"
)

type SegmentRootMapping struct {
	WorkPackageHash types.WorkPackageHash
	SegmentRoot     types.OpaqueHash
}

type CE134WorkPackageShare struct {
	CoreIndex           types.CoreIndex
	SegmentRootMappings []SegmentRootMapping
	Bundle              []byte
}

// Helper to decode segment-root mappings from the stream.
// Count uses GP len++ (compact) encoding per spec.
func readSegmentRootMappings(r io.Reader) ([]SegmentRootMapping, error) {
	var mappings []SegmentRootMapping
	count, err := readCompactLength(r)
	if err != nil {
		return nil, err
	}
	for i := uint64(0); i < count; i++ {
		var wpHash types.WorkPackageHash
		var segRoot types.OpaqueHash
		if _, err := io.ReadFull(r, wpHash[:]); err != nil {
			return nil, err
		}
		if _, err := io.ReadFull(r, segRoot[:]); err != nil {
			return nil, err
		}
		mappings = append(mappings, SegmentRootMapping{wpHash, segRoot})
	}
	return mappings, nil
}

// Handler for CE 134: Guarantor <-> Guarantor work-package sharing
// Accepts any io.ReadWriteCloser for testability, and allows injection of a PVMExecutor for unit tests.
func HandleWorkPackageShare(
	_ blockchain.Blockchain,
	stream io.ReadWriteCloser,
	keypair keystore.KeyPair,
	pvmExecutor work_package.PVMExecutor,
) error {
	// 1. Read core index (2 bytes, little endian)
	coreIndexBuf := make([]byte, 2)
	if _, err := io.ReadFull(stream, coreIndexBuf); err != nil {
		return err
	}
	coreIndex := types.CoreIndex(binary.LittleEndian.Uint16(coreIndexBuf))

	// 2. Read segment-root mappings
	_, err := readSegmentRootMappings(stream)
	if err != nil {
		return fmt.Errorf("failed to read segment-root mappings: %w", err)
	}
	// (Note: mappings are not used in this minimal handler; add logic as needed)

	// 3. Read work-package bundle (rest of stream until FIN)
	bundle := make([]byte, 65536)
	n, err := stream.Read(bundle)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read bundle: %w", err)
	}
	bundle = bundle[:n]

	// 4. Basic verification: decode bundle, check authorization, check mappings
	controller := work_package.NewSharedController(bundle, coreIndex)
	if pvmExecutor != nil {
		controller.PVM = pvmExecutor
	}
	workReport, err := controller.Process()
	if err != nil {
		return fmt.Errorf("work-package verification failed: %w", err)
	}

	workReportSerialization := utilities.WorkReportSerialization(workReport)
	workReportHash := hash.Blake2bHash(workReportSerialization)

	// message: JamGuarantee context + work report hash
	message := []byte(types.JamGuarantee)
	message = append(message, workReportHash[:]...)

	sig, err := keypair.Sign(message)
	if err != nil {
		return fmt.Errorf("failed to sign work-report hash: %w", err)
	}

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

	coreIndexBytes := encodeLE16(uint16(workPackage.CoreIndex))

	// Get WorkPackage bytes using ScaleEncode
	wpBytes, err := workPackage.WorkPackage.ScaleEncode()
	if err != nil {
		return nil, fmt.Errorf("failed to encode WorkPackage: %w", err)
	}

	// CE134 spec: Core Index ++ Segments-Root Mappings (Core Index = u16); no HeaderHash
	totalLen := 1 + 2 + len(wpBytes)

	result := make([]byte, 0, totalLen)

	result = append(result, requestType)
	result = append(result, coreIndexBytes...)
	result = append(result, wpBytes...)

	return result, nil
}
