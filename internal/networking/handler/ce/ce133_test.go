package ce

import (
	"encoding/binary"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
)

func makeCE133FirstMsg(coreIndex uint32, workPackage []byte) []byte {
	buf := make([]byte, 2+len(workPackage))
	binary.LittleEndian.PutUint16(buf[:2], uint16(coreIndex))
	copy(buf[2:], workPackage)
	return buf
}

func TestHandleWorkPackageSubmission_Basic(t *testing.T) {
	coreIndex := uint32(7)
	workPackage := []byte{0xAA, 0xBB, 0xCC}
	extrinsics := []byte{0x11, 0x22, 0x33, 0x44}

	firstMsg := makeCE133FirstMsg(coreIndex, workPackage)
	stream := newMockStream(firstMsg)
	stream.r.Write(extrinsics)

	err := HandleWorkPackageSubmission(nil, &quic.Stream{Stream: stream})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	resp := stream.w.Bytes()
	// CE133 spec: response is FIN only (stream close); no message bytes
	if len(resp) != 0 {
		t.Errorf("expected no response bytes (FIN only), got %x", resp)
	}
}

func TestHandleWorkPackageSubmission_Minimal(t *testing.T) {
	coreIndex := uint32(0)
	workPackage := []byte{}
	extrinsics := []byte{}

	firstMsg := makeCE133FirstMsg(coreIndex, workPackage)
	stream := newMockStream(firstMsg)
	stream.r.Write(extrinsics)

	err := HandleWorkPackageSubmission(nil, &quic.Stream{Stream: stream})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	resp := stream.w.Bytes()
	// CE133 spec: response is FIN only (stream close); no message bytes
	if len(resp) != 0 {
		t.Errorf("expected no response bytes (FIN only), got %x", resp)
	}
}
