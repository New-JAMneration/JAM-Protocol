package ce

import (
	"encoding/binary"
	"os"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
)

func TestHandleAuditShardRequest(t *testing.T) {
	os.Setenv("USE_MINI_REDIS", "true")
	defer store.CloseMiniRedis()

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
	msgLen := binary.LittleEndian.Uint32(resp[:4])
	if 4+msgLen > uint32(len(resp)) {
		t.Fatalf("response truncated")
	}
	payload := resp[4 : 4+msgLen]

	justificationStart := -1
	var justification []byte
	for i := 0; i < len(payload); i++ {
		if payload[i] == 0x00 {
			if i+33 <= len(payload) {
				justificationStart = i
				justification = payload[i:]
				break
			}
		}
	}

	if justificationStart == -1 {
		t.Fatalf("could not find justification start (0x00 discriminator)")
	}

	if len(justification) < 1 {
		t.Fatalf("justification is too short")
	}

	if justification[0] != 0x00 {
		t.Fatalf("justification does not start with expected discriminator, got %x", justification[0])
	}

	if len(justification) < 33 {
		t.Fatalf("justification is too short, expected at least 33 bytes, got %d", len(justification))
	}

	t.Logf("Success: Found justification starting at position %d, length %d bytes", justificationStart, len(justification))
}
