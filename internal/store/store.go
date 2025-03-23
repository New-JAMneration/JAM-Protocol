package store

import (
	"log"
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

var (
	initOnce    sync.Once
	globalStore *Store
)

// Store represents a thread-safe global state container
type Store struct {
	mu sync.RWMutex

	// INFO: Add more fields here
	unfinalizedBlocks           *UnfinalizedBlocks
	processingBlock             *ProcessingBlock
	priorStates                 *PriorStates
	intermediateStates          *IntermediateStates
	posteriorStates             *PosteriorStates
	ancestorHeaders             *AncestorHeaders
	intermediateHeader          *IntermediateHeader
	posteriorCurrentValidators  *PosteriorCurrentValidators
	beefyCommitmentOutput       *BeefyCommitmentOutputs   // This is tmp used waiting for more def in GP
	accumulatedWorkReports      *AccumulatedWorkReports   // W^! (accumulated immediately)
	queuedWorkReports           *QueuedWorkReports        // W^Q (queued execution)
	accumulatableWorkReports    *AccumulatableWorkReports // W^* (accumulatable work-reports in this block)
	accumulationStatistics      *AccumulationStatistics
	deferredTransfersStatistics *DeferredTransfersStatistics
}

// GetInstance returns the singleton instance of Store.
// If the instance doesn't exist, it creates one.
func GetInstance() *Store {
	initOnce.Do(func() {
		globalStore = &Store{
			unfinalizedBlocks:           NewUnfinalizedBlocks(),
			processingBlock:             NewProcessingBlock(),
			priorStates:                 NewPriorStates(),
			intermediateStates:          NewIntermediateStates(),
			posteriorStates:             NewPosteriorStates(),
			ancestorHeaders:             NewAncestorHeaders(),
			intermediateHeader:          NewIntermediateHeader(),
			posteriorCurrentValidators:  NewPosteriorValidators(),
			beefyCommitmentOutput:       NewBeefyCommitmentOutput(), // This is tmp used waiting for more def in GP
			accumulatedWorkReports:      NewAccumulatedWorkReports(),
			queuedWorkReports:           NewQueuedWorkReports(),
			accumulatableWorkReports:    NewAccumulatableWorkReports(),
			accumulationStatistics:      NewAccumulationStatistics(),
			deferredTransfersStatistics: NewDeferredTransfersStatistics(),
		}
		log.Println("🚀 Store initialized")
	})
	return globalStore
}

func (s *Store) AddBlock(block types.Block) {
	s.unfinalizedBlocks.AddBlock(block)
}

func (s *Store) GetBlocks() []types.Block {
	return s.unfinalizedBlocks.GetAllAncientBlocks()
}

func (s *Store) GetBlock() types.Block {
	return s.unfinalizedBlocks.GetLatestBlock()
}

func (s *Store) GetLatestBlock() types.Block {
	return s.unfinalizedBlocks.GetLatestBlock()
}

func (s *Store) GetProcessingBlockPointer() *ProcessingBlock {
	return s.processingBlock
}

func (s *Store) GenerateGenesisBlock(block types.Block) {
	s.unfinalizedBlocks.GenerateGenesisBlock(block)
	s.unfinalizedBlocks.AddBlock(block)
	log.Println("🚀 Genesis block generated")
}

// Set
func (s *Store) GetPriorStates() *PriorStates {
	return s.priorStates
}

func (s *Store) GetIntermediateStates() *IntermediateStates {
	return s.intermediateStates
}

func (s *Store) GetIntermediateHeaders() *IntermediateHeader {
	return s.intermediateHeader
}

func (s *Store) GetPosteriorStates() *PosteriorStates {
	return s.posteriorStates
}

func (s *Store) GenerateGenesisState(state types.State) {
	s.priorStates.GenerateGenesisState(state)
	log.Println("🚀 Genesis state generated")
}

// AncestorHeaders

func (s *Store) AddAncestorHeader(header types.Header) {
	s.ancestorHeaders.AddHeader(header)
}

func (s *Store) GetAncestorHeaders() []types.Header {
	return s.ancestorHeaders.GetHeaders()
}

// PosteriorCurrentValidators

func (s *Store) AddPosteriorCurrentValidator(validator types.Validator) {
	s.posteriorCurrentValidators.AddValidator(validator)
}

func (s *Store) GetPosteriorCurrentValidators() types.ValidatorsData {
	return s.posteriorCurrentValidators.GetValidators()
}

func (s *Store) GetPosteriorCurrentValidatorByIndex(index types.ValidatorIndex) types.Validator {
	return s.posteriorCurrentValidators.GetValidatorByIndex(index)
}

// IntermediateHeader

func (s *Store) AddIntermediateHeader(header types.Header) {
	s.intermediateHeader.AddHeader(header)
}

func (s *Store) GetIntermediateHeader() types.Header {
	return s.intermediateHeader.GetHeader()
}

func (s *Store) GetIntermediateHeaderPointer() *IntermediateHeader {
	return s.intermediateHeader
}

func (s *Store) ResetIntermediateHeader() {
	s.intermediateHeader.ResetHeader()
}

// BeefyCommitmentOutput (This is tmp used waiting for more def in GP)

// Get
func (s *Store) GetBeefyCommitmentOutput() types.BeefyCommitmentOutput {
	return s.beefyCommitmentOutput.GetBeefyCommitmentOutput()
}

// Set
func (s *Store) GetBeefyCommitmentOutputs() *BeefyCommitmentOutputs {
	return s.beefyCommitmentOutput
}

// // ServiceAccountDerivatives (This is tmp used waiting for more testvector to verify)

// // Get
// func (s *Store) GetServiceAccountDerivatives() types.ServiceAccountDerivatives {
// 	return s.serviceAccountDerivatives.GetServiceAccountDerivatives()
// }

// // Set
// func (s *Store) GetServiceAccountDerivatives() *ServiceAccountDerivatives {
// 	return s.serviceAccountDerivatives
// }

// AccumulatedWorkReports
func (s *Store) GetAccumulatedWorkReportsPointer() *AccumulatedWorkReports {
	return s.accumulatedWorkReports
}

// QueuedWorkReports
func (s *Store) GetQueuedWorkReportsPointer() *QueuedWorkReports {
	return s.queuedWorkReports
}

// AccumulatableWorkReports
func (s *Store) GetAccumulatableWorkReportsPointer() *AccumulatableWorkReports {
	return s.accumulatableWorkReports
}

// AccumulationStatistics
func (s *Store) GetAccumulationStatisticsPointer() *AccumulationStatistics {
	return s.accumulationStatistics
}

// DeferredTransfersStatistics
func (s *Store) GetDeferredTransfersStatisticsPointer() *DeferredTransfersStatistics {
	return s.deferredTransfersStatistics
}
