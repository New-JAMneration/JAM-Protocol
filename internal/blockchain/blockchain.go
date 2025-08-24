package blockchain

import (
	"log"

	"github.com/New-JAMneration/JAM-Protocol/internal/db"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// Just the template for the interface
// TODO: Implement the Blockchain interface.
// Blockchain defines the required interface for retrieving blocks.
type BlockchainApi interface {
	// GetBlockNumber returns the block number for the given block hash.
	GetBlockNumber(types.HeaderHash) (uint32, error)
	// GetBlockHashByNumber returns candidate block hashes for the specified block number.
	GetBlockHashByNumber(number uint32) ([]types.HeaderHash, error)
	// GetBlock returns a block for the given block hash.
	GetBlock(types.HeaderHash) (types.Block, error)
	// GenesisBlockHash returns the genesis block hash.
	GenesisBlockHash() types.HeaderHash
}

type State struct {
	priorStates        *PriorStates
	intermediateStates *IntermediateStates
	posteriorStates    *PosteriorStates
}

type Blockchain struct {
	db db.KeyValueDB

	genesisBlock               *types.Block
	unfinalizedBlocks          *UnfinalizedBlocks
	finalizedIndex             map[types.HeaderHash]bool
	processingBlock            *ProcessingBlock
	priorStates                *PriorStates
	ancestorHeaders            *AncestorHeaders
	posteriorCurrentValidators *PosteriorCurrentValidators

	state   *State
	statedb *StateDB
}

func (bc *Blockchain) AddBlock(block types.Block) {
	bc.unfinalizedBlocks.AddBlock(block)
}

func (bc *Blockchain) GetBlocks() []types.Block {
	return bc.unfinalizedBlocks.GetAllAncientBlocks()
}

func (bc *Blockchain) GetBlock() types.Block {
	return bc.unfinalizedBlocks.GetLatestBlock()
}

func (bc *Blockchain) GetLatestBlock() types.Block {
	return bc.unfinalizedBlocks.GetLatestBlock()
}

func (bc *Blockchain) GetProcessingBlockPointer() *ProcessingBlock {
	return bc.processingBlock
}

func (bc *Blockchain) GenerateGenesisBlock(block types.Block) {
	bc.unfinalizedBlocks.GenerateGenesisBlock(block)
	bc.unfinalizedBlocks.AddBlock(block)
	// Genesis block is always finalized
	bc.finalizedIndex[block.Header.Parent] = true
}

// Finalized Blocks Management

// FinalizeBlock marks a block as finalized by its hash
func (bc *Blockchain) FinalizeBlock(blockHash types.HeaderHash) {
	bc.finalizedIndex[blockHash] = true
}

// IsBlockFinalized checks if a block is finalized
func (bc *Blockchain) IsBlockFinalized(blockHash types.HeaderHash) bool {
	return bc.finalizedIndex[blockHash]
}

// GetFinalizedBlocks returns all finalized blocks
// TODO: This is not efficient,
// We’re checking the blocks in reverse order—if we encounter a finalized block,
// can we assume that all remaining blocks are also finalized?
func (bc *Blockchain) GetFinalizedBlocks() []types.Block {
	allBlocks := bc.unfinalizedBlocks.GetAllAncientBlocks()
	var finalizedBlocks []types.Block

	for i := len(allBlocks) - 1; i >= 0; i-- {
		block := allBlocks[i]
		if bc.IsBlockFinalized(block.Header.Parent) {
			finalizedBlocks = append(finalizedBlocks, block)
		}
	}

	return finalizedBlocks
}

// GetUnfinalizedBlocks returns all unfinalized blocks
func (bc *Blockchain) GetUnfinalizedBlocks() []types.Block {
	allBlocks := bc.unfinalizedBlocks.GetAllAncientBlocks()
	var unfinalizedBlocks []types.Block

	for i := len(allBlocks) - 1; i >= 0; i-- {
		block := allBlocks[i]
		if bc.IsBlockFinalized(block.Header.Parent) {
			break
		}
		unfinalizedBlocks = append(unfinalizedBlocks, block)
	}

	return unfinalizedBlocks
}

// GetLatestFinalizedBlock returns the most recent finalized block
func (bc *Blockchain) GetLatestFinalizedBlock() types.Block {
	allBlocks := bc.unfinalizedBlocks.GetAllAncientBlocks()

	// Search from the end to find the latest finalized block
	for i := len(allBlocks) - 1; i >= 0; i-- {
		if bc.IsBlockFinalized(allBlocks[i].Header.Parent) {
			return allBlocks[i]
		}
	}

	return types.Block{}
}

// CleanupOldFinalizedBlocks removes old finalized blocks from memory
// This is a simple implementation - you might want to implement more sophisticated cleanup
func (bc *Blockchain) CleanupOldFinalizedBlocks(keepCount int) {
	// TODO: Implement this
}

// Set
func (bc *Blockchain) GetPriorStates() *PriorStates {
	return bc.priorStates
}

func (bc *Blockchain) GetIntermediateStates() *IntermediateStates {
	return bc.state.intermediateStates
}

func (bc *Blockchain) GetPosteriorStates() *PosteriorStates {
	return bc.state.posteriorStates
}

func (bc *Blockchain) GenerateGenesisState(state types.State) {
	bc.priorStates.GenerateGenesisState(state)
	log.Println("🚀 Genesis state generated")
}

// AncestorHeaders

func (bc *Blockchain) AddAncestorHeader(header types.Header) {
	bc.ancestorHeaders.AddHeader(header)
}

func (bc *Blockchain) GetAncestorHeaders() []types.Header {
	return bc.ancestorHeaders.GetHeaders()
}

// PosteriorCurrentValidators

func (bc *Blockchain) AddPosteriorCurrentValidator(validator types.Validator) {
	bc.posteriorCurrentValidators.AddValidator(validator)
}

func (bc *Blockchain) GetPosteriorCurrentValidators() types.ValidatorsData {
	return bc.posteriorCurrentValidators.GetValidators()
}

func (bc *Blockchain) GetPosteriorCurrentValidatorByIndex(index types.ValidatorIndex) types.Validator {
	return bc.posteriorCurrentValidators.GetValidatorByIndex(index)
}

// // ServiceAccountDerivatives (This is tmp used waiting for more testvector to verify)

// // Get
// func (bc *Blockchain) GetServiceAccountDerivatives() typebc.ServiceAccountDerivatives {
// 	return bc.serviceAccountDerivativebc.GetServiceAccountDerivatives()
// }

// // Set
// func (bc *Blockchain) GetServiceAccountDerivatives() *ServiceAccountDerivatives {
// 	return s.serviceAccountDerivatives
// }
