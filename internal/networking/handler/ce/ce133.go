package ce

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type CE133WorkPackageSubmission struct {
	CoreIndex   uint32
	WorkPackage []byte
	Extrinsics  []byte
}

// [TODO-Validation]
// 1. Validate extrinsic data is correctly covering all the extrinsic referenced by work-package.
// 2. Reject extrinsics which contain the imported segments.
func HandleWorkPackageSubmission(blockchain blockchain.Blockchain, stream *quic.Stream) error {
	// Read first message: 4 bytes core index + work-package (rest of message)
	firstMsg := make([]byte, 4096) // Arbitrary max size for demo; adjust as needed
	n, err := stream.Read(firstMsg)
	if err != nil && err != io.EOF {
		return err
	}
	if n < 4 {
		return io.ErrUnexpectedEOF
	}
	coreIndex := binary.LittleEndian.Uint32(firstMsg[:4])
	workPackage := make([]byte, n-4)
	copy(workPackage, firstMsg[4:n])

	// Read second message: all extrinsic data (until FIN)
	extra := make([]byte, 65536)
	exLen, err := stream.Read(extra)
	if err != nil && err != io.EOF {
		return err
	}
	extrinsics := make([]byte, exLen)
	copy(extrinsics, extra[:exLen])

	_ = CE133WorkPackageSubmission{
		CoreIndex:   coreIndex,
		WorkPackage: workPackage,
		Extrinsics:  extrinsics,
	}
	stream.Write([]byte{0x01})
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

	encoder := types.NewEncoder()

	if err := h.writeBytes(encoder, encodeLE32(workpackage.CoreIndex)); err != nil {

		return nil, fmt.Errorf("failed to encode CoreIndex for WorkPackageSubmission: %w", err)
	}

	if err := h.writeBytes(encoder, workpackage.WorkPackage); err != nil {
		return nil, fmt.Errorf("failed to encode WorkPackage for WorkPackageSubmission: %w", err)
	}
	if err := h.writeBytes(encoder, workpackage.Extrinsics); err != nil {
		return nil, fmt.Errorf("failed to encode Extrinsics for WorkPackageSubmission: %w", err)
	}

	totalLen := 4 + len(workpackage.WorkPackage) + len(workpackage.Extrinsics)
	result := make([]byte, 0, totalLen)
	result = append(result, encodeLE32(workpackage.CoreIndex)...)
	result = append(result, workpackage.WorkPackage...)
	result = append(result, workpackage.Extrinsics...)
	return result, nil
}
