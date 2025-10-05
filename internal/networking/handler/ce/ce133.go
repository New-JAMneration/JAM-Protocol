package ce

import (
	"encoding/binary"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type CE133WorkPackageSubmission struct {
	CoreIndex   uint16
	WorkPackage []byte
	Extrinsics  []byte
}

// HandleWorkPackageSubmission_Builder handles work package submission from Builder role
func HandleWorkPackageSubmission_Builder(blockchain blockchain.Blockchain, stream *quic.Stream) error {
	lenBuf := make([]byte, 4)
	if err := stream.WriteMessage(lenBuf); err != nil {
		return fmt.Errorf("failed to read first message length: %w", err)
	}
	firstMsgLen := binary.LittleEndian.Uint32(lenBuf)

	firstMsg := make([]byte, firstMsgLen)
	if err := stream.WriteMessage(firstMsg); err != nil {
		return fmt.Errorf("failed to read first message content: %w", err)
	}

	if firstMsgLen < 2 {
		return fmt.Errorf("first message too short, expected at least 2 bytes for core index")
	}

	// Extract Core Index (2 bytes) and Work-Package
	coreIndex := binary.LittleEndian.Uint16(firstMsg[:2])
	workPackage := firstMsg[2:]
	_ = coreIndex // reserved for validation/routing

	var wp types.WorkPackage
	decoder := types.NewDecoder()
	if err := decoder.Decode(workPackage, &wp); err != nil {
		return fmt.Errorf("failed to decode work package: %w", err)
	}
	if err := wp.Validate(); err != nil {
		return fmt.Errorf("invalid work package: %w", err)
	}

	if err := stream.WriteMessage(lenBuf); err != nil {
		return fmt.Errorf("failed to read extrinsics length: %w", err)
	}
	extrinsicsLen := binary.LittleEndian.Uint32(lenBuf)

	// Validate extrinsic message size equals the sum of declared extrinsic lengths in the WorkPackage
	expectedSize, err := expectedExtrinsicsSize(&wp)
	if err != nil {
		return err
	} else if extrinsicsLen != expectedSize {
		return fmt.Errorf("bad_extrinsics_message_size: got %d, want %d", extrinsicsLen, expectedSize)
	}

	extrinsics := make([]byte, extrinsicsLen)
	if err := stream.ReadFull(extrinsics); err != nil {
		return fmt.Errorf("failed to read extrinsics data: %w", err)
	}

	finBuf := make([]byte, 3)
	if err := stream.WriteMessage(finBuf); err != nil {
		return fmt.Errorf("failed to read FIN marker: %w", err)
	}
	if string(finBuf) != "FIN" {
		return fmt.Errorf("expected FIN marker, got %q", string(finBuf))
	}

	if err := stream.ReadFull([]byte("FIN")); err != nil {
		return fmt.Errorf("failed to write FIN response: %w", err)
	}

	return stream.Close()
}

// HandleWorkPackageSubmission_Guarantor handles work package submission from Guarantor role
//
// [TODO-Validation]
// 1. Validate extrinsic data is correctly covering all the extrinsic referenced by work-package.
// 2. Reject extrinsics which contain the imported segments.
func HandleWorkPackageSubmission_Guarantor(blockchain blockchain.Blockchain, stream *quic.Stream) error {
	// Read first message: Core Index (2 bytes) + Work-Package
	lenBuf := make([]byte, 4)
	if err := stream.ReadFull(lenBuf); err != nil {
		return fmt.Errorf("failed to read first message length: %w", err)
	}
	firstMsgLen := binary.LittleEndian.Uint32(lenBuf)

	firstMsg := make([]byte, firstMsgLen)
	if err := stream.ReadFull(firstMsg); err != nil {
		return fmt.Errorf("failed to read first message content: %w", err)
	} else if firstMsgLen < 2 {
		return fmt.Errorf("first message too short, expected at least 2 bytes for core index")
	}

	// Extract Core Index (2 bytes) and Work-Package
	coreIndex := binary.LittleEndian.Uint16(firstMsg[:2])
	workPackage := firstMsg[2:]
	_ = coreIndex // reserved for validation/routing

	// Decode and validate Work-Package
	var wp types.WorkPackage
	decoder := types.NewDecoder()
	if err := decoder.Decode(workPackage, &wp); err != nil {
		return fmt.Errorf("failed to decode work package: %w", err)
	}
	if err := wp.Validate(); err != nil {
		return fmt.Errorf("invalid work package: %w", err)
	}

	// Read second message: Extrinsic data array
	if err := stream.ReadFull(lenBuf); err != nil {
		return fmt.Errorf("failed to read extrinsics length: %w", err)
	}
	extrinsicsLen := binary.LittleEndian.Uint32(lenBuf)

	// Validate extrinsic message size equals the sum of declared extrinsic lengths in the WorkPackage
	expectedSize, err := expectedExtrinsicsSize(&wp)
	if err != nil {
		return err
	} else if extrinsicsLen != expectedSize {
		return fmt.Errorf("bad_extrinsics_message_size: got %d, want %d", extrinsicsLen, expectedSize)
	}

	// Read extrinsic data payload
	extrinsics := make([]byte, extrinsicsLen)
	if err := stream.ReadFull(extrinsics); err != nil {
		return fmt.Errorf("failed to read extrinsics data: %w", err)
	}

	finBuf := make([]byte, 3)
	if err := stream.ReadFull(finBuf); err != nil {
		return fmt.Errorf("failed to read FIN marker: %w", err)
	}
	if string(finBuf) != "FIN" {
		return fmt.Errorf("expected FIN marker, got %q", string(finBuf))
	}

	if _, err := stream.Write([]byte("FIN")); err != nil {
		return fmt.Errorf("failed to write FIN response: %w", err)
	}

	return stream.Close()
}

func (h *DefaultCERequestHandler) encodeWorkPackageSubmission(message interface{}) ([]byte, error) {

	workpackage, ok := message.(*CE133WorkPackageSubmission)
	if !ok {
		return nil, fmt.Errorf("unsupported message type for WorkPackageSubmission: %T", message)
	}
	if workpackage == nil {
		return nil, fmt.Errorf("nil payload for WorkPackageSubmission")
	}

	// Build payload once: CoreIndex (2 bytes) ++ WorkPackage ++ Extrinsics
	totalLen := 2 + len(workpackage.WorkPackage) + len(workpackage.Extrinsics)
	result := make([]byte, 0, totalLen)
	result = append(result, encodeLE16(workpackage.CoreIndex)...)
	result = append(result, workpackage.WorkPackage...)
	result = append(result, workpackage.Extrinsics...)
	return result, nil
}

func expectedExtrinsicsSize(wp *types.WorkPackage) (uint32, error) {
	var sum uint64
	for _, item := range wp.Items {
		for _, ex := range item.Extrinsic {
			sum += uint64(ex.Len)
		}
	}
	if sum > uint64(^uint32(0)) {
		return 0, fmt.Errorf("extrinsics_size_overflow: %d", sum)
	}
	return uint32(sum), nil
}
