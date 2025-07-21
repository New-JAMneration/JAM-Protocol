package ce

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestHandleWorkReportRequest_Basic(t *testing.T) {
	// Create a minimal WorkReport and compute its hash (use zeroed hash for test)
	wp := &types.WorkReport{}
	var hash types.WorkReportHash // zeroed for test

	// Set up a fake lookup map
	lookupMap := map[types.WorkReportHash]*types.WorkReport{
		hash: wp,
	}
	lookup := func(h types.WorkReportHash) (*types.WorkReport, bool) {
		wr, ok := lookupMap[h]
		return wr, ok
	}

	// Prepare the request: 32-byte hash + 'FIN'
	input := append(hash[:], []byte("FIN")...)
	stream := newMockStream(input)

	err := HandleWorkReportRequest(stream, lookup)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	// Check the response: should be the encoded work-report + 'FIN'
	resp := stream.w.Bytes()
	if len(resp) < 3 || string(resp[len(resp)-3:]) != "FIN" {
		t.Fatalf("expected response to end with FIN, got %x", resp)
	}

	data := resp[:len(resp)-3]
	decoder := types.NewDecoder()
	var got types.WorkReport
	if err := decoder.Decode(data, &got); err != nil {
		t.Fatalf("failed to decode work-report: %v", err)
	}
}
