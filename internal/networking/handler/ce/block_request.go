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

// BlockRequest represents a block request message.
// It consists of a 32-byte header hash, a 1-byte direction, and a 4-byte maximum block count.
type BlockRequest struct {
	HeaderHash types.HeaderHash // Reference block hash
	Direction  byte             // 0: Ascending exclusive, 1: Descending inclusive
	MaxBlocks  uint32           // Maximum number of blocks requested
}

func HandleBlockRequest(blockchain blockchain.Blockchain, req BlockRequest) ([]*types.Block, error) {
	count := req.MaxBlocks
	var blocks []*types.Block

	log.Printf("Handling block request for %v, direction %d, count %d\n", req.HeaderHash, req.Direction, count)

	switch req.Direction {
	case 0: // Ascending exclusive: start with a child of the given block.
		startTimeSlot, err := blockchain.GetBlockTimeSlot(req.HeaderHash)
		if err != nil {
			return nil, err
		}
		currentHash := req.HeaderHash
		for i := uint32(0); i < count; i++ {
			candidateTimeSlot := startTimeSlot + types.TimeSlot(i)
			candidateHashes, err := blockchain.GetBlockHashByTimeSlot(candidateTimeSlot)
			if err != nil {
				break
			}
			found := false
			for _, candidate := range candidateHashes {
				blk, err := blockchain.GetBlockByHash(candidate)
				if err != nil {
					continue
				}
				// Check if the candidate's parent matches the current block.
				if blk.Header.Parent == currentHash {
					blocks = append(blocks, blk)
					currentHash = candidate
					found = true
					break
				}
			}
			if !found {
				break
			}
		}
	case 1: // Descending inclusive: start with the given block and traverse to ancestors.
		currentHash := req.HeaderHash
		for i := uint32(0); i < count; i++ {
			blk, err := blockchain.GetBlockByHash(currentHash)
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
	// Ensure the payload length is at least 37 bytes.
	reqPayload, err := io.ReadAll(stream)
	if err != nil {
		return err
	}
	if reqPayload[0] < 32+1+4 {
		return errors.New("invalid block request length")
	}

	reqPayload = reqPayload[4:]

	var req BlockRequest
	copy(req.HeaderHash[:], reqPayload[:32])
	req.Direction = reqPayload[32]

	fmt.Printf("req: %v\n", reqPayload[33:37])
	req.MaxBlocks = binary.LittleEndian.Uint32(reqPayload[33:37])

	// Process the block request.
	blocks, err := HandleBlockRequest(blockchain, req)
	if err != nil {
		return err
	}

	fmt.Printf("number of blocks: %v\n", len(blocks))

	// Encode and write the response blocks.
	encoder := types.NewEncoder()

	// Write each block as a framed message using our custom encoding.
	for _, blk := range blocks {

		blkData, err := encoder.Encode(&blk)
		if err != nil {
			log.Printf("failed to encode block: %v", err)
			continue
		}

		// Write the length prefix.
		sizeBuf := make([]byte, 4)
		binary.LittleEndian.PutUint32(sizeBuf, uint32(len(blkData)))
		if _, err := stream.Write(sizeBuf); err != nil {
			log.Printf("failed to write length prefix: %v", err)
			return err
		}
		// Write the encoded block data.
		if _, err := stream.Write(blkData); err != nil {
			log.Printf("failed to write block data: %v", err)
			return err
		}

		log.Printf("Writing block %v\n", blk.Header)
	}

	return stream.Close()
}
