package ce

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
)

func makeCE133FirstMsg(coreIndex uint32, workPackage []byte) []byte {
	buf := make([]byte, 4+len(workPackage))
	binary.LittleEndian.PutUint32(buf[:4], coreIndex)
	copy(buf[4:], workPackage)
	return buf
}

func TestHandleWorkPackageSubmission_Basic(t *testing.T) {
	coreIndex := uint32(7)
	workPackage := []byte{0xAA, 0xBB, 0xCC}
	extrinsics := []byte{0x11, 0x22, 0x33, 0x44}

	firstMsg := makeCE133FirstMsg(coreIndex, workPackage)
	stream := newMockStream(firstMsg)
	stream.r.Write(extrinsics)

	err := HandleWorkPackageSubmission_Builder(nil, &quic.Stream{Stream: stream})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	resp := stream.w.Bytes()
	if !bytes.Contains(resp, []byte{0x01}) {
		t.Errorf("expected handler to write response 0xEF, got %x", resp)
	}
}

func TestHandleWorkPackageSubmission_Minimal(t *testing.T) {
	coreIndex := uint32(0)
	workPackage := []byte{}
	extrinsics := []byte{}

	firstMsg := makeCE133FirstMsg(coreIndex, workPackage)
	stream := newMockStream(firstMsg)
	stream.r.Write(extrinsics)

	err := HandleWorkPackageSubmission_Builder(nil, &quic.Stream{Stream: stream})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	resp := stream.w.Bytes()
	if !bytes.Contains(resp, []byte{0x01}) {
		t.Errorf("expected handler to write response 0xEF, got %x", resp)
	}
}
