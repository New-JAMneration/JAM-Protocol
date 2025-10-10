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
	"github.com/New-JAMneration/JAM-Protocol/internal/safrole"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type CE131Payload struct {
	EpochIndex uint32
	Attempt    uint8     // 0 or 1 (single byte)
	Proof      [784]byte // RingVRF
}

var localBandersnatchKey types.BandersnatchPublic

// SetLocalBandersnatchKey allows tests to inject the local validator's Bandersnatch key.
func SetLocalBandersnatchKey(key types.BandersnatchPublic) {
	localBandersnatchKey = key
}

// Role: [Validator -> Validator]
//
// [TODO-Validation]
// 1. Check finality of the block for stopping forwarding.
// 2. Check finality is running behind the state to reset or stop the stream.
func HandleSafroleTicketDistribution(blockchain blockchain.Blockchain, stream *quic.Stream) error {
	payload := make([]byte, 789) // Epoch Index ++ Ticket
	if err := stream.ReadFull(payload); err != nil {
		return err
	}
	var req CE131Payload
	req.EpochIndex = binary.LittleEndian.Uint32(payload[:4])
	req.Attempt = payload[4]
	copy(req.Proof[:], payload[5:789])

	// calculate proxy validator index
	proxyIndexBytes := req.Proof[780:784]
	nextValidators := store.GetInstance().GetPosteriorStates().GetGammaK()
	if len(nextValidators) == 0 {
		return stream.Close()
	}
	proxyIndex := binary.BigEndian.Uint32(proxyIndexBytes) % uint32(len(nextValidators))
	proxyValidator := nextValidators[proxyIndex]

	// Proxy validator must verify the ticket proof
	if err := verifySafroleTicketProof(req); err != nil {
		return fmt.Errorf("RingVRF proof verification failed: %w", err)
	}

	if localBandersnatchKey != proxyValidator.Bandersnatch {
		delaySlots := int(math.Max(float64(types.EpochLength)/60.0, 1.0))
		time.Sleep(time.Duration(delaySlots) * time.Duration(types.SlotPeriod) * time.Second)

		// CE131 send ticket to the proxy validator
		addr := string(bytes.TrimRight(proxyValidator.Metadata[:32], "\x00"))
		if addr == "" {
			return fmt.Errorf("invalid proxy validator network address in metadata")
		}

		if err := stream.WriteMessage(payload); err != nil {
			return fmt.Errorf("failed to write payload: %w", err)
		}

		fin := make([]byte, 3)
		if err := stream.ReadFull(fin); err != nil {
			return fmt.Errorf("failed to read FIN response: %w", err)
		} else if string(fin) != "FIN" {
			return fmt.Errorf("unexpected FIN response: %q", string(fin))
		}
		return stream.Close()
	}

	// As proxy (CE132): forward to all current validators
	// Forwarding should be delayed until max(floor(E/20), 1) slots after connectivity changes,
	// and evenly spaced out until half-way through the Safrole lottery period.
	delaySlots := int(math.Max(float64(types.EpochLength)/20.0, 1.0))

	lotteryPeriod := types.EpochLength - types.SlotSubmissionEnd
	halfLotteryPeriod := lotteryPeriod / 2

	remainingSlots := halfLotteryPeriod - delaySlots + types.SlotSubmissionEnd
	currentValidators := store.GetInstance().GetPosteriorStates().GetKappa()
	if len(currentValidators) == 0 {
		if err := stream.WriteMessage([]byte("FIN")); err != nil {
			return fmt.Errorf("failed to write FIN response: %w", err)
		}
		return stream.Close()
	}

	// Sleep until the initial delay threshold
	time.Sleep(time.Duration(delaySlots) * time.Duration(types.SlotPeriod) * time.Second)

	// Compute batching: send to multiple validators per slot.
	batches := int(math.Max(1.0, float64(remainingSlots)))
	batchSize := int(math.Ceil(float64(len(currentValidators)) / float64(batches)))

	go func(validators types.ValidatorsData) {
		for start := 0; start < len(validators); start += batchSize {
			end := min(start+batchSize, len(validators))

			// Send within this slot to the batch without inter-send sleep
			for _, validator := range validators[start:end] {
				if err := forwardSafroleTicket(validator, payload); err != nil {
					fmt.Printf("Failed to forward ticket to validator (Ed25519: %x): %v\n", validator.Ed25519[:8], err)
				}
			}

			// Sleep exactly one slot between batches, except after the last batch
			if end < len(validators) {
				time.Sleep(time.Duration(types.SlotPeriod) * time.Second)
			}
		}
	}(currentValidators)

	if err := stream.WriteMessage([]byte("FIN")); err != nil {
		return fmt.Errorf("failed to write FIN response: %w", err)
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

	protocolID := []byte{132}
	if _, err := stream.Write(protocolID); err != nil {
		return fmt.Errorf("failed to write protocol ID: %w", err)
	}

	if _, err := stream.Write(payload); err != nil {
		return fmt.Errorf("failed to write payload: %w", err)
	}

	ack := make([]byte, 1)
	if _, err := stream.Read(ack); err != nil {
		return fmt.Errorf("failed to read acknowledgment: %w", err)
	} else if ack[0] != 0x01 {
		return fmt.Errorf("received invalid acknowledgment: %x", ack[0])
	}

	return nil
}

func verifySafroleTicketProof(req CE131Payload) error {
	ticketEnvelope := types.TicketEnvelope{
		Attempt:   types.TicketAttempt(req.Attempt),
		Signature: types.BandersnatchRingVrfSignature(req.Proof),
	}

	ticketsExtrinsic := types.TicketsExtrinsic{ticketEnvelope}

	_, err := safrole.VerifyTicketsProof(ticketsExtrinsic)
	if err != nil {
		return fmt.Errorf("ticket proof verification failed: %w", err)
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

	result := make([]byte, 0, 789) // 4 + 1 + 784 bytes
	result = append(result, encodeLE32(ticketDist.EpochIndex)...)
	result = append(result, ticketDist.Attempt)
	result = append(result, ticketDist.Proof[:]...)
	return result, nil
}
