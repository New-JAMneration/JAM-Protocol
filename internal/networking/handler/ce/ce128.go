package ce

import (
	"encoding/binary"
	"errors"
	"fmt"
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
	reqPayload, err := stream.ReadMessage()
	if err != nil {
		return err
	}
	if len(reqPayload) < CE128MinRequestSize {
		return errors.New("invalid block request length")
	}

	var req CE128Payload
	copy(req.HeaderHash[:], reqPayload[:HashSize])
	req.Direction = reqPayload[HashSize]
	req.MaxBlocks = binary.LittleEndian.Uint32(reqPayload[HashSize+1 : HashSize+1+U32Size])

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
		if err := stream.WriteMessage(blkData); err != nil {
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

	// Spec: Header Hash ++ Direction ++ Maximum Blocks (single message, no double encode)
	result := make([]byte, 0, HashSize+1+U32Size)
	result = append(result, blockReq.HeaderHash[:]...)
	result = append(result, blockReq.Direction)
	result = append(result, encodeLE32(blockReq.MaxBlocks)...)
	return result, nil
}
