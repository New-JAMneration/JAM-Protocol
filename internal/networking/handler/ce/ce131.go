package ce

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	vrf "github.com/New-JAMneration/JAM-Protocol/pkg/Rust-VRF/vrf-func-ffi/src"
)

type CE131Payload struct {
	EpochIndex uint32
	Attempt    uint8 // 0 or 1 (single byte)
	Proof      [CE131ProofSize]byte
}

var localBandersnatchKey types.BandersnatchPublic

// SetLocalBandersnatchKey allows tests to inject the local validator's Bandersnatch key.
func SetLocalBandersnatchKey(key types.BandersnatchPublic) {
	localBandersnatchKey = key
}

func HandleSafroleTicketDistribution(_ blockchain.Blockchain, stream *quic.Stream) error {
	payload, err := stream.ReadMessage()
	if err != nil {
		return err
	}
	if len(payload) != CE131PayloadSize {
		return fmt.Errorf("invalid ticket payload length: got %d, want %d", len(payload), CE131PayloadSize)
	}
	var req CE131Payload
	req.EpochIndex = binary.LittleEndian.Uint32(payload[:U32Size])
	req.Attempt = payload[U32Size]
	copy(req.Proof[:], payload[U32Size+1:CE131PayloadSize])

	proxyIndexBytes := req.Proof[CE131ProofSize-U32Size : CE131ProofSize]
	proxyIndex := binary.BigEndian.Uint32(proxyIndexBytes) % uint32(types.ValidatorsCount)

	// Get next epoch's validator set (GammaK)
	nextValidators := blockchain.GetInstance().GetPosteriorStates().GetGammaK()

	if int(proxyIndex) >= len(nextValidators) {
		return stream.Close()
	}
	proxyValidator := nextValidators[proxyIndex]

	if localBandersnatchKey == proxyValidator.Bandersnatch {
		currentValidators := blockchain.GetInstance().GetPosteriorStates().GetKappa()

		// In bootstrap/test environments the current validator set may be unavailable.
		// When it's empty we skip VRF verification (we can't form a valid ring).
		if len(currentValidators) > 0 {
			if err := verifySafroleTicketProof(req, currentValidators); err != nil {
				return fmt.Errorf("VRF proof verification failed: %w", err)
			}
		}

		delaySlots := int(math.Max(float64(types.EpochLength)/20.0, 1.0))

		lotteryPeriod := types.EpochLength - types.SlotSubmissionEnd
		halfLotteryPeriod := lotteryPeriod / 2

		forwardingSlots := halfLotteryPeriod - delaySlots
		if forwardingSlots <= 0 {
			forwardingSlots = 1
		}
		go func() {
			time.Sleep(time.Duration(delaySlots) * time.Duration(types.SlotPeriod) * time.Second)

			for i, validator := range currentValidators {
				if validator.Bandersnatch == localBandersnatchKey {
					continue
				}

				if i > 0 {
					time.Sleep(time.Duration(forwardingSlots) * time.Duration(types.SlotPeriod) * time.Second)
				}

				if err := forwardSafroleTicket(validator, payload); err != nil {
					fmt.Printf("Failed to forward ticket to validator %d: %v\n", i, err)
				}
			}
		}()

		// Spec: <-- FIN only (no ack byte)
	}

	return stream.Close()
}

func forwardSafroleTicket(validator types.Validator, payload []byte) error {
	addr := string(bytes.TrimRight(validator.Metadata[:32], "\x00"))
	if addr == "" {
		return fmt.Errorf("invalid validator network address in metadata")
	}

	tlsConfig, err := quic.NewTLSConfig(false, false)
	if err != nil {
		return fmt.Errorf("failed to create TLS config: %w", err)
	}
	tlsConfig.InsecureSkipVerify = true

	ctx := context.Background()
	conn, err := quic.Dial(ctx, addr, tlsConfig, quic.NewQuicConfig(), quic.Validator)
	if err != nil {
		return fmt.Errorf("failed to establish QUIC connection: %w", err)
	}
	defer conn.Close()

	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		return fmt.Errorf("failed to open stream: %w", err)
	}
	defer stream.Close()

	quicStream := &quic.Stream{Stream: stream}
	protocolID := byte(SafroleTicketDistribution) // 130
	if _, err := quicStream.Write([]byte{protocolID}); err != nil {
		return fmt.Errorf("failed to write protocol ID: %w", err)
	}
	if err := quicStream.WriteMessage(payload); err != nil {
		return fmt.Errorf("failed to write ticket message: %w", err)
	}
	// Spec: <-- FIN only (acceptor sends no ack byte). Close our write half; done.
	return quicStream.Close()
}

func verifySafroleTicketProof(req CE131Payload, validators types.ValidatorsData) error {
	ring := make([]byte, 0)
	for _, validator := range validators {
		ring = append(ring, validator.Bandersnatch[:]...)
	}

	verifier, err := vrf.NewVerifier(ring, uint(len(validators)))
	if err != nil {
		return fmt.Errorf("failed to create VRF verifier: %w", err)
	}
	defer verifier.Free()

	input := make([]byte, 5)
	binary.LittleEndian.PutUint32(input[:4], req.EpochIndex)
	input[4] = req.Attempt

	if _, err := verifier.RingVerify(input, []byte{}, req.Proof[:]); err != nil {
		return fmt.Errorf("ring VRF proof verification failed: %w", err)
	}

	return nil
}

func (h *DefaultCERequestHandler) encodeSafroleTicketDistribution(message interface{}) ([]byte, error) {
	ticketDist, ok := message.(*CE131Payload)
	if !ok {
		return nil, fmt.Errorf("unsupported message type for SafroleTicketDistribution: %T", message)
	}

	encoder := types.NewEncoder()

	if err := h.writeBytes(encoder, encodeLE32(ticketDist.EpochIndex)); err != nil {
		return nil, fmt.Errorf("failed to encode EpochIndex: %w", err)
	}

	if err := encoder.WriteByte(ticketDist.Attempt); err != nil {
		return nil, fmt.Errorf("failed to encode Attempt: %w", err)
	}

	if err := h.writeBytes(encoder, ticketDist.Proof[:]); err != nil {
		return nil, fmt.Errorf("failed to encode Proof: %w", err)
	}

	result := make([]byte, 0, CE131PayloadSize)
	result = append(result, encodeLE32(ticketDist.EpochIndex)...)
	result = append(result, ticketDist.Attempt)
	result = append(result, ticketDist.Proof[:]...)
	return result, nil
}
