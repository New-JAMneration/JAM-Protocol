package ce

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math/bits"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/keystore"
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

// readCompactLength reads a GP len++ compact integer from an io.Reader.
func readCompactLength(r io.Reader) (uint64, error) {
	prefix := make([]byte, 1)
	if _, err := io.ReadFull(r, prefix); err != nil {
		return 0, err
	}
	l := bits.LeadingZeros8(^prefix[0])
	if l > 0 {
		extra := make([]byte, l)
		if _, err := io.ReadFull(r, extra); err != nil {
			return 0, err
		}
		prefix = append(prefix, extra...)
	}
	return types.NewDecoder().DecodeUint(prefix)
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

// ce134Stream is the stream interface for CE134: supports JAMNP message framing (ReadMessage, WriteMessage).
type ce134Stream interface {
	io.ReadWriteCloser
	ReadMessage() ([]byte, error)
	WriteMessage(payload []byte) error
}

// Handler for CE 134: Guarantor <-> Guarantor work-package sharing
// Message 1: Core Index ++ Segments-Root Mappings; Message 2: Work-Package Bundle.
func HandleWorkPackageShare(
	_ blockchain.Blockchain,
	stream ce134Stream,
	keypair keystore.KeyPair,
	pvmExecutor work_package.PVMExecutor,
) error {
	// 1. Read first framed message: Core Index (2 bytes) + Segment-Root Mappings
	msg1, err := stream.ReadMessage()
	if err != nil {
		return fmt.Errorf("failed to read core index and mappings: %w", err)
	}
	if len(msg1) < 2 {
		return fmt.Errorf("first message too short for core index")
	}
	coreIndex := types.CoreIndex(binary.LittleEndian.Uint16(msg1[:2]))

	// 2. Parse segment-root mappings from the rest of message 1
	_, err = readSegmentRootMappings(bytes.NewReader(msg1[2:]))
	if err != nil {
		return fmt.Errorf("failed to read segment-root mappings: %w", err)
	}

	// 3. Read second framed message: Work-Package Bundle
	bundle, err := stream.ReadMessage()
	if err != nil {
		return fmt.Errorf("failed to read bundle: %w", err)
	}

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
	if err := stream.WriteMessage(resp); err != nil {
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

	// Segments-Root Mappings: len++ [(WorkPackageHash ++ SegmentRoot)...]
	mappings := workPackage.SegmentRootMappings
	if mappings == nil {
		mappings = []SegmentRootMapping{}
	}
	enc := types.NewEncoder()
	mappingsLenBytes, err := enc.EncodeUint(uint64(len(mappings)))
	if err != nil {
		return nil, fmt.Errorf("failed to encode mappings length: %w", err)
	}

	wpBytes, err := workPackage.WorkPackage.ScaleEncode()
	if err != nil {
		return nil, fmt.Errorf("failed to encode WorkPackage: %w", err)
	}

	result := make([]byte, 0, 1+2+len(mappingsLenBytes)+len(mappings)*(32+32)+len(wpBytes))
	result = append(result, requestType)
	result = append(result, coreIndexBytes...)
	result = append(result, mappingsLenBytes...)
	for _, m := range mappings {
		result = append(result, m.WorkPackageHash[:]...)
		result = append(result, m.SegmentRoot[:]...)
	}
	result = append(result, wpBytes...)

	return result, nil
}
