package ce

import (
	"encoding/binary"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/database/provider/memory"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
)

func TestHandleAuditShardRequest(t *testing.T) {
	db := memory.NewDatabase()
	defer db.Close()
	SetDatabase(db)
	defer SetDatabase(nil)

	erasureRoot := []byte("fake-erasure-root-32bytes-long!!")
	shardIndex := uint32(5)

	// Prepare request: one length-prefixed message = erasureRoot (HashSize) + shardIndex (U16Size)
	req := make([]byte, 0, CE138RequestSize)
	req = append(req, erasureRoot...)
	req = append(req, byte(shardIndex), byte(shardIndex>>8))
	stream := newMockStream(framePayload(req))

	err := HandleAuditShardRequest(nil, &quic.Stream{Stream: stream})
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	resp := stream.w.Bytes()
	if len(resp) < 4 {
		t.Fatalf("response too short")
	}
	// First framed message: bundle shard
	msg1Len := binary.LittleEndian.Uint32(resp[:4])
	if 4+msg1Len > uint32(len(resp)) {
		t.Fatalf("first message truncated")
	}
	_ = resp[4 : 4+msg1Len] // bundle shard

	// Second framed message: justification
	rest := resp[4+msg1Len:]
	if len(rest) < 4 {
		t.Fatalf("response too short for second message")
	}
	msg2Len := binary.LittleEndian.Uint32(rest[:4])
	if 4+msg2Len > uint32(len(rest)) {
		t.Fatalf("second message truncated")
	}
	justification := rest[4 : 4+msg2Len]

	if len(justification) < 1 {
		t.Fatalf("justification is too short")
	}

	if justification[0] != 0x00 {
		t.Fatalf("justification does not start with expected discriminator, got %x", justification[0])
	}

	if len(justification) < 33 {
		t.Fatalf("justification is too short, expected at least 33 bytes, got %d", len(justification))
	}

	t.Logf("Success: bundle shard %d bytes, justification %d bytes", msg1Len, msg2Len)
}
