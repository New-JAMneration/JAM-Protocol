package ce

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
)

// HandleBundleRequest handles CE147: an auditor's request for a full work-package bundle from a guarantor.
//
// Request (Auditor -> Guarantor):
//
//	Erasure-Root (32 bytes)
//	FIN
//
// Response (Guarantor -> Auditor):
//
//	Work-Package Bundle (JAM codec bytes)
//	FIN
//
// If the guarantor cannot supply a valid bundle, the auditor should fall back to CE138 per JAMNP.
func HandleBundleRequest(bc blockchain.Blockchain, stream *quic.Stream) error {
	payload, err := stream.ReadMessage()
	if err != nil {
		return err
	}
	if len(payload) != CE147RequestSize {
		return fmt.Errorf("bundle request: expected %d-byte erasure root, got %d", CE147RequestSize, len(payload))
	}
	erasureRoot := payload

	bundleBytes, err := GetKV(DB(bc), wpBundleKey(erasureRoot))
	if err != nil {
		return fmt.Errorf("get work-package bundle: %w", err)
	}
	if len(bundleBytes) == 0 {
		return fmt.Errorf("work-package bundle not found for erasure root")
	}

	if err := stream.WriteMessage(bundleBytes); err != nil {
		return err
	}
	return stream.Close()
}

// CE147Payload is the client-side request body for CE147 (after the protocol ID byte).
type CE147Payload struct {
	ErasureRoot []byte
}

func (h *DefaultCERequestHandler) encodeBundleRequest(message interface{}) ([]byte, error) {
	req, ok := message.(*CE147Payload)
	if !ok {
		return nil, fmt.Errorf("unsupported message type for BundleRequest: %T", message)
	}
	if req == nil {
		return nil, fmt.Errorf("nil payload for BundleRequest")
	}
	if len(req.ErasureRoot) != HashSize {
		return nil, fmt.Errorf("erasure root must be exactly %d bytes, got %d", HashSize, len(req.ErasureRoot))
	}
	out := make([]byte, CE147RequestSize)
	copy(out, req.ErasureRoot)
	return out, nil
}
