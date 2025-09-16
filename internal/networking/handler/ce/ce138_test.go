package ce

import (
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

	// Prepare request: erasureRoot (32 bytes) + shardIndex (4 bytes LE) + 'FIN'
	req := make([]byte, 0, 32+4+3)
	req = append(req, erasureRoot...)
	req = append(req, byte(shardIndex), byte(shardIndex>>8), byte(shardIndex>>16), byte(shardIndex>>24))
	req = append(req, []byte("FIN")...)
	stream := newMockStream(req)

	err := HandleAuditShardRequest_Assurer(nil, &quic.Stream{Stream: stream})
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	resp := stream.w.Bytes()
	if len(resp) < 3 || string(resp[len(resp)-3:]) != "FIN" {
		t.Fatalf("expected response to end with FIN, got %x", resp)
	}

	resp = resp[:len(resp)-3] // remove FIN

	if len(resp) == 0 {
		t.Fatalf("response is empty")
	}

	justificationStart := -1
	for i := 0; i < len(resp); i++ {
		if resp[i] == 0x00 {
			if i+33 <= len(resp) {
				justificationStart = i
				break
			}
		}
	}

	if justificationStart == -1 {
		t.Fatalf("could not find justification start (0x00 discriminator)")
	}

	justification := resp[justificationStart:]
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
