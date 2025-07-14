package ce

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func makeCE131Payload(epoch uint32, attempt uint8, proof [784]byte) []byte {
	buf := make([]byte, 789)
	binary.LittleEndian.PutUint32(buf[:4], epoch)
	buf[4] = attempt
	copy(buf[5:], proof[:])
	return buf
}

func TestHandleSafroleTicketDistribution_Proxy(t *testing.T) {
	// Setup: 2 validators, local is proxy
	var v0, v1 types.Validator
	v0.Bandersnatch = types.BandersnatchPublic{1}
	v1.Bandersnatch = types.BandersnatchPublic{2}
	validators := types.ValidatorsData{v0, v1}
	store.GetInstance().GetPosteriorStates().SetGammaK(validators)

	SetLocalBandersnatchKey(v1.Bandersnatch)

	// Craft proof so proxyIndex = 1
	var proof [784]byte
	binary.BigEndian.PutUint32(proof[780:784], 1)
	payload := makeCE131Payload(42, 0, proof)
	stream := newMockStream(payload)

	err := HandleSafroleTicketDistribution(nil, &quic.Stream{Stream: stream})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	resp := stream.w.Bytes()
	if !bytes.Contains(resp, []byte{0xAB}) {
		t.Errorf("expected proxy to write response, got %x", resp)
	}
}

func TestHandleSafroleTicketDistribution_NotProxy(t *testing.T) {
	// Setup: 2 validators, local is not proxy
	var v0, v1 types.Validator
	v0.Bandersnatch = types.BandersnatchPublic{1}
	v1.Bandersnatch = types.BandersnatchPublic{2}
	validators := types.ValidatorsData{v0, v1}
	store.GetInstance().GetPosteriorStates().SetGammaK(validators)

	SetLocalBandersnatchKey(v0.Bandersnatch)

	// Craft proof so proxyIndex = 1
	var proof [784]byte
	binary.BigEndian.PutUint32(proof[780:784], 1)
	payload := makeCE131Payload(42, 0, proof)
	stream := newMockStream(payload)

	err := HandleSafroleTicketDistribution(nil, &quic.Stream{Stream: stream})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	resp := stream.w.Bytes()
	if bytes.Contains(resp, []byte{0xAB}) {
		t.Errorf("expected no response for non-proxy, got %x", resp)
	}
}
