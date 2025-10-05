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
	payload := make([]byte, 789)
	if err := stream.ReadFull(payload); err != nil {
		return err
	}
	var req CE131Payload
	req.EpochIndex = binary.LittleEndian.Uint32(payload[:4])
	req.Attempt = payload[4]
	copy(req.Proof[:], payload[5:789])

	// calculate proxy validator index
	proxyIndexBytes := req.Proof[780:784]
	proxyIndex := binary.BigEndian.Uint32(proxyIndexBytes) % uint32(types.ValidatorsCount)
	nextValidators := store.GetInstance().GetPosteriorStates().GetGammaK()

	if int(proxyIndex) >= len(nextValidators) {
		return stream.Close()
	}
	// the proxy validator is selected from next epoch validators
	proxyValidator := nextValidators[proxyIndex]

	// CE131
	delaySlots := 0
	if localBandersnatchKey != proxyValidator.Bandersnatch {
		// max(|E/60|, 1) slots, E is the number of slots in an epoch
		delaySlots = int(math.Max(float64(types.SlotSubmissionEnd)/60.0, 1.0))
		time.Sleep(time.Duration(delaySlots) * time.Duration(types.SlotPeriod) * time.Second)
	}

	// CE132: proxy validator to all current validators
	// if so, the validator is proxy validator, we should skip first step (generate proxy validator)
	// and do second step.
	if err := verifySafroleTicketProof(req); err != nil {
		return fmt.Errorf("VRF proof verification failed: %w", err)
	}

	lotteryPeriod := types.EpochLength - types.SlotSubmissionEnd
	halfLotteryPeriod := lotteryPeriod / 2

	// start forwarding after half-way through the Safrole lottery period.
	forwardingSlots := halfLotteryPeriod - delaySlots
	if forwardingSlots <= 0 {
		forwardingSlots = 1
	}
	time.Sleep(time.Duration(forwardingSlots) * time.Duration(types.SlotPeriod) * time.Second)
	go func() {
		currentValidators := store.GetInstance().GetPosteriorStates().GetKappa()
		for _, validator := range currentValidators {
			if err := forwardSafroleTicket(validator, payload); err != nil {
				fmt.Printf("Failed to forward ticket to validator %d: %v\n", validator, err)
			}
		}
	}()

	if _, err := stream.Write([]byte("FIN")); err != nil {
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
