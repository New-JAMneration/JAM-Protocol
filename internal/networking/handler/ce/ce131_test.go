package ce

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func makeCE131Payload(epoch uint32, attempt uint8, proof [CE131ProofSize]byte) []byte {
	buf := make([]byte, CE131PayloadSize)
	binary.LittleEndian.PutUint32(buf[:U32Size], epoch)
	buf[U32Size] = attempt
	copy(buf[U32Size+1:], proof[:])
	return buf
}

// makeFramedCE131Payload returns a JAMNP-framed message (4-byte LE length + CE131PayloadSize ticket).
func makeFramedCE131Payload(epoch uint32, attempt uint8, proof [CE131ProofSize]byte) []byte {
	payload := makeCE131Payload(epoch, attempt, proof)
	framed := make([]byte, 0, 4+len(payload))
	framed = binary.LittleEndian.AppendUint32(framed, uint32(len(payload)))
	framed = append(framed, payload...)
	return framed
}

func TestHandleSafroleTicketDistribution_Proxy(t *testing.T) {
	// Setup: 2 validators, local is proxy
	var v0, v1 types.Validator
	v0.Bandersnatch = types.BandersnatchPublic{1}
	v1.Bandersnatch = types.BandersnatchPublic{2}
	validators := types.ValidatorsData{v0, v1}
	blockchain.GetInstance().GetPosteriorStates().SetGammaK(validators)

	SetLocalBandersnatchKey(v1.Bandersnatch)

	// Craft proof so proxyIndex = 1
	var proof [CE131ProofSize]byte
	binary.BigEndian.PutUint32(proof[CE131ProofSize-U32Size:CE131ProofSize], 1)
	framed := makeFramedCE131Payload(42, 0, proof)
	stream := newMockStream(framed)

	err := HandleSafroleTicketDistribution(nil, &quic.Stream{Stream: stream})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	resp := stream.w.Bytes()
	// Spec: <-- FIN only (no ack byte). Proxy just closes the stream.
	if len(resp) != 0 {
		t.Errorf("expected no response bytes (FIN only), got %x", resp)
	}
}

func TestHandleSafroleTicketDistribution_NotProxy(t *testing.T) {
	// Setup: 2 validators, local is not proxy
	var v0, v1 types.Validator
	v0.Bandersnatch = types.BandersnatchPublic{1}
	v1.Bandersnatch = types.BandersnatchPublic{2}
	validators := types.ValidatorsData{v0, v1}
	blockchain.GetInstance().GetPosteriorStates().SetGammaK(validators)

	SetLocalBandersnatchKey(v0.Bandersnatch)

	// Craft proof so proxyIndex = 1
	var proof [CE131ProofSize]byte
	binary.BigEndian.PutUint32(proof[CE131ProofSize-U32Size:CE131ProofSize], 1)
	framed := makeFramedCE131Payload(42, 0, proof)
	stream := newMockStream(framed)

	err := HandleSafroleTicketDistribution(nil, &quic.Stream{Stream: stream})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	resp := stream.w.Bytes()
	// Spec: non-proxy also sends <-- FIN only; no 0x01 ack in either case.
	if bytes.Contains(resp, []byte{0x01}) {
		t.Errorf("expected no ack byte, got %x", resp)
	}
}
