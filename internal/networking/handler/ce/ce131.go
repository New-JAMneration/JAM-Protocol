package ce

import (
	"encoding/binary"
	"io"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
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

func HandleSafroleTicketDistribution(blockchain blockchain.Blockchain, stream *quic.Stream) error {
	payload := make([]byte, 789)
	if _, err := io.ReadFull(stream, payload); err != nil {
		return err
	}
	var req CE131Payload
	req.EpochIndex = binary.LittleEndian.Uint32(payload[:4])
	req.Attempt = payload[4]
	copy(req.Proof[:], payload[5:789])

	// Extract proxy index from last 4 bytes of proof (big-endian)
	proxyIndexBytes := req.Proof[780:784]
	proxyIndex := binary.BigEndian.Uint32(proxyIndexBytes) % uint32(types.ValidatorsCount)

	// Get next epoch's validator set (GammaK)
	nextValidators := store.GetInstance().GetPosteriorStates().GetGammaK()

	// Defensive: check bounds
	if int(proxyIndex) >= len(nextValidators) {
		return stream.Close()
	}
	proxyValidator := nextValidators[proxyIndex]

	// If this node is the proxy, write a response (for testability)
	if localBandersnatchKey == proxyValidator.Bandersnatch {
		// For test: write a single byte to indicate success
		stream.Write([]byte{0x01})
	}

	return stream.Close()
}
