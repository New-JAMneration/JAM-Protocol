package up

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// BuildHandshake builds the local UP 0 handshake from known blocks and finalized hash.
func BuildHandshake(blocks []types.Block, finalized types.HeaderHash) (Handshake, error) {
	cv, finalRef, err := ViewAtFinalized(blocks, finalized)
	if err != nil {
		return Handshake{}, err
	}
	return Handshake{
		Final:  finalRef,
		Leaves: cv.CollectLeaves(finalized),
	}, nil
}

// WriteHandshake builds, encodes, and writes the local handshake.
func WriteHandshake(stream framedStream, blocks []types.Block, finalized types.HeaderHash) (Handshake, error) {
	hs, err := BuildHandshake(blocks, finalized)
	if err != nil {
		return Handshake{}, err
	}
	payload, err := EncodeHandshake(hs)
	if err != nil {
		return Handshake{}, err
	}
	if err := stream.WriteMessage(payload); err != nil {
		return Handshake{}, err
	}
	return hs, nil
}

// HandshakeRefs returns final and all leaves for announcement tracking.
func HandshakeRefs(h Handshake) []BlockRef {
	refs := make([]BlockRef, 0, 1+len(h.Leaves))
	refs = append(refs, h.Final)
	return append(refs, h.Leaves...)
}
