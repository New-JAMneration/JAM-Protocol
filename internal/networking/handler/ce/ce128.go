package ce

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// CE128Payload represents a block request message.
// It consists of a 32-byte header hash, a 1-byte direction, and a 4-byte maximum block count.
type CE128Payload struct {
	HeaderHash types.HeaderHash // Reference block hash
	Direction  byte             // 0: Ascending exclusive, 1: Descending inclusive
	MaxBlocks  uint32           // Maximum number of blocks requested
}

func HandleBlockRequest(blockchain blockchain.Blockchain, req CE128Payload) ([]types.Block, error) {
	count := req.MaxBlocks
	var blocks []types.Block

	log.Printf("Handling block request for %v, direction %d, count %d\n", req.HeaderHash, req.Direction, count)

	switch req.Direction {
	case 0:
		currentHash := req.HeaderHash
		for i := uint32(0); i < count; i++ {
			found := false
			for blockNum := uint32(0); blockNum < 10; blockNum++ {
				candidateHashes, err := blockchain.GetBlockHashByNumber(blockNum)
				if err != nil {
					continue
				}
				for _, candidate := range candidateHashes {
					blk, err := blockchain.GetBlock(candidate)
					if err != nil {
						continue
					}
					log.Printf("[CE128] i=%d, blockNum=%d, currentHash=%x, candidate=%x, blk.Header.Parent=%x", i, blockNum, currentHash, candidate, blk.Header.Parent)
					if blk.Header.Parent == currentHash && candidate != currentHash {
						log.Printf("[CE128] MATCH: candidate=%x", candidate)
						blocks = append(blocks, blk)
						currentHash = candidate
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if !found {
				log.Printf("[CE128] No match found for currentHash=%x at i=%d", currentHash, i)
				break
			}
		}
	case 1: // Descending inclusive: start with the given block and traverse to ancestors.
		currentHash := req.HeaderHash
		for i := uint32(0); i < count; i++ {
			blk, err := blockchain.GetBlock(currentHash)
			if err != nil {
				break
			}
			blocks = append(blocks, blk)
			// If we've reached the genesis block, stop.
			if currentHash == blockchain.GenesisBlockHash() {
				break
			}
			currentHash = blk.Header.Parent
		}
	default:
		return nil, errors.New("invalid direction")
	}
	return blocks, nil
}

func HandleBlockRequestStream(blockchain blockchain.Blockchain, stream *quic.Stream) error {
	// The quic.DefaultCEHandler has already read the protocol ID (1 byte).
	// Now we need to read the remaining payload from the stream.
	reqPayload, err := io.ReadAll(stream)
	if err != nil {
		return err
	}

	// The payload should be at least 37 bytes (32 + 1 + 4)
	if len(reqPayload) < 32+1+4 {
		return errors.New("invalid block request length")
	}

	var req CE128Payload
	copy(req.HeaderHash[:], reqPayload[:32])
	req.Direction = reqPayload[32]
	req.MaxBlocks = binary.LittleEndian.Uint32(reqPayload[33:37])

	blocks, err := HandleBlockRequest(blockchain, req)
	if err != nil {
		return err
	}

	fmt.Printf("number of blocks: %v\n", len(blocks))

	encoder := types.NewEncoder()

	for _, blk := range blocks {

		blkData, err := encoder.Encode(&blk)
		if err != nil {
			log.Printf("failed to encode block: %v", err)
			continue
		}

		sizeBuf := make([]byte, 4)
		binary.LittleEndian.PutUint32(sizeBuf, uint32(len(blkData)))
		if _, err := stream.Write(sizeBuf); err != nil {
			log.Printf("failed to write length prefix: %v", err)
			return err
		}
		if _, err := stream.Write(blkData); err != nil {
			log.Printf("failed to write block data: %v", err)
			return err
		}

		log.Printf("Writing block %v\n", blk.Header)
	}

	return stream.Close()
}

func (p *CE128Payload) Encode(e *types.Encoder) error {
	if err := p.HeaderHash.Encode(e); err != nil {
		return err
	}
	if err := e.WriteByte(p.Direction); err != nil {
		return err
	}
	maxBlocks := types.U32(p.MaxBlocks)
	if err := maxBlocks.Encode(e); err != nil {
		return err
	}
	return nil
}

func (h *DefaultCERequestHandler) encodeBlockRequest(message interface{}) ([]byte, error) {
	blockReq, ok := message.(*CE128Payload)
	if !ok {
		return nil, fmt.Errorf("unsupported message type for BlockRequest: %T", message)
	}

	h.encoder = types.NewEncoder()

	if err := blockReq.HeaderHash.Encode(h.encoder); err != nil {
		return nil, fmt.Errorf("failed to encode HeaderHash: %w", err)
	}

	if err := h.encoder.WriteByte(blockReq.Direction); err != nil {
		return nil, fmt.Errorf("failed to encode Direction: %w", err)
	}

	maxBlocks := types.U32(blockReq.MaxBlocks)
	if err := maxBlocks.Encode(h.encoder); err != nil {
		return nil, fmt.Errorf("failed to encode MaxBlocks: %w", err)
	}

	encoded, err := h.encoder.Encode(blockReq)
	if err != nil {
		return nil, fmt.Errorf("failed to encode CE128Payload: %w", err)
	}
	return encoded, nil
}
