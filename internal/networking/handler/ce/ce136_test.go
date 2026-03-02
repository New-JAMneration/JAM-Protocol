package ce

import (
	"encoding/binary"
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

	input := hash[:]
	stream := newMockStream(input)

	err := HandleWorkReportRequest(stream, lookup)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	resp := stream.w.Bytes()
	if len(resp) < 4 {
		t.Fatalf("response too short for message frame")
	}
	n := binary.LittleEndian.Uint32(resp[:4])
	payload := resp[4:]
	if uint32(len(payload)) < n {
		t.Fatalf("response truncated: want %d payload bytes, got %d", n, len(payload))
	}
	payload = payload[:n]
	decoder := types.NewDecoder()
	var got types.WorkReport
	if err := decoder.Decode(payload, &got); err != nil {
		t.Fatalf("failed to decode work-report: %v", err)
	}
}
