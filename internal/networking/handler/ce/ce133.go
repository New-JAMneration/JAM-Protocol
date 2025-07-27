package ce

import (
	"encoding/binary"
	"io"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
)

type CE133WorkPackageSubmission struct {
	CoreIndex   uint32
	WorkPackage []byte
	Extrinsics  []byte
}

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
