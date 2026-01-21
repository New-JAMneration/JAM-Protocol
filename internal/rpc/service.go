package rpc

import (
	"encoding/base64"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/New-JAMneration/JAM-Protocol/logger"
)

type RPCService struct {
	chainState *blockchain.ChainState
}

func NewRPCService() *RPCService {
	return &RPCService{
		chainState: blockchain.GetInstance(),
	}
}

func (s *RPCService) Parameters() (*ChainParameters, error) {
	snapshot := types.SnapshotProtocolParams()

	return &ChainParameters{
		V1: ChainParametersV1{
			DepositPerItem:        snapshot.BI,
			DepositPerByte:        snapshot.BL,
			DepositPerAccount:     snapshot.BS,
			CoreCount:             uint32(snapshot.C),
			MinTurnaroundPeriod:   snapshot.D,
			EpochPeriod:           snapshot.E,
			MaxAccumulateGas:      snapshot.GA,
			MaxIsAuthorizedGas:    snapshot.GI,
			MaxRefineGas:          snapshot.GR,
			BlockGasLimit:         snapshot.GT,
			RecentBlockCount:      uint32(snapshot.H),
			MaxWorkItems:          uint32(snapshot.I),
			MaxDependencies:       uint32(snapshot.J),
			MaxTicketsPerBlock:    uint32(snapshot.K),
			MaxLookupAnchorAge:    snapshot.L,
			TicketsAttemptsNumber: uint32(snapshot.N),
			AuthWindow:            uint32(snapshot.O),
			SlotPeriodSec:         uint32(snapshot.P),
			AuthQueueLen:          uint32(snapshot.Q),
			RotationPeriod:        uint32(snapshot.R),
			MaxExtrinsics:         uint32(snapshot.T),
			AvailabilityTimeout:   uint32(snapshot.U),
			ValCount:              uint32(snapshot.V),
			MaxAuthorizerCodeSize: snapshot.WA,
			MaxInput:              snapshot.WB,
			MaxServiceCodeSize:    snapshot.WC,
			BasicPieceLen:         snapshot.WE,
			MaxImports:            snapshot.WM,
			SegmentPieceCount:     snapshot.WP,
			MaxReportElectiveData: snapshot.WR,
			TransferMemoSize:      snapshot.WT,
			MaxExports:            snapshot.WX,
			EpochTailStart:        snapshot.Y,
		},
	}, nil
}

func (s *RPCService) BestBlock() (*BlockDescriptor, error) {
	latestBlock := s.chainState.GetLatestBlock()

	headerHash, err := hash.ComputeBlockHeaderHash(latestBlock.Header)
	if err != nil {
		return nil, fmt.Errorf("failed to compute block header hash: %w", err)
	}

	return &BlockDescriptor{
		HeaderHash: encodeHash(headerHash),
		Slot:       uint64(latestBlock.Header.Slot),
	}, nil
}

func (s *RPCService) FinalizedBlock() (*BlockDescriptor, error) {
	finalizedBlocks := s.chainState.GetFinalizedBlocks()
	if len(finalizedBlocks) == 0 {
		// if no finalized block, return genesis block
		genesisBlock := s.chainState.GetGenesisBlock()
		headerHash, err := hash.ComputeBlockHeaderHash(genesisBlock.Header)
		if err != nil {
			return nil, fmt.Errorf("failed to compute genesis block header hash: %w", err)
		}
		return &BlockDescriptor{
			HeaderHash: encodeHash(headerHash),
			Slot:       uint64(genesisBlock.Header.Slot),
		}, nil
	}

	// get last finalized block
	lastFinalized := finalizedBlocks[len(finalizedBlocks)-1]
	headerHash, err := hash.ComputeBlockHeaderHash(lastFinalized.Header)
	if err != nil {
		return nil, fmt.Errorf("failed to compute finalized block header hash: %w", err)
	}

	return &BlockDescriptor{
		HeaderHash: encodeHash(headerHash),
		Slot:       uint64(lastFinalized.Header.Slot),
	}, nil
}

func (s *RPCService) Parent(headerHashStr string) (*BlockDescriptor, error) {
	headerHash, err := decodeHash(headerHashStr)
	if err != nil {
		return nil, fmt.Errorf("invalid header hash: %w", err)
	}

	block, err := s.chainState.GetBlockByHash(headerHash)
	if err != nil {
		return nil, NewRPCErrorWithData(ErrCodeBlockUnavailable, "block unavailable", headerHashStr)
	}

	return &BlockDescriptor{
		HeaderHash: encodeHash(block.Header.Parent),
		Slot:       uint64(block.Header.Slot - 1),
	}, nil
}

func (s *RPCService) StateRoot(headerHashStr string) (string, error) {
	headerHash, err := decodeHash(headerHashStr)
	if err != nil {
		return "", fmt.Errorf("invalid header hash: %w", err)
	}

	block, err := s.chainState.GetBlockByHash(headerHash)
	if err != nil {
		return "", NewRPCErrorWithData(ErrCodeBlockUnavailable, "block unavailable", headerHashStr)
	}

	// TODO: get state root from posterior state
	return encodeHash(block.Header.ParentStateRoot), nil
}

func (s *RPCService) BeefyRoot(headerHashStr string) (string, error) {
	headerHash, err := decodeHash(headerHashStr)
	if err != nil {
		return "", fmt.Errorf("invalid header hash: %w", err)
	}

	_, err = s.chainState.GetBlockByHash(headerHash)
	if err != nil {
		return "", NewRPCErrorWithData(ErrCodeBlockUnavailable, "block unavailable", headerHashStr)
	}

	// TODO: implement get BEEFY root
	emptyHash := types.HeaderHash{}
	return encodeHash(emptyHash), nil
}

func (s *RPCService) Statistics(headerHashStr string) (string, error) {
	headerHash, err := decodeHash(headerHashStr)
	if err != nil {
		return "", fmt.Errorf("invalid header hash: %w", err)
	}

	_, err = s.chainState.GetBlockByHash(headerHash)
	if err != nil {
		return "", NewRPCErrorWithData(ErrCodeBlockUnavailable, "block unavailable", headerHashStr)
	}

	// TODO: implement get block statistics from posterior state
	return encodeBlob([]byte{}), nil
}

func (s *RPCService) ServiceData(headerHashStr string, serviceID uint64) (string, error) {
	headerHash, err := decodeHash(headerHashStr)
	if err != nil {
		return "", fmt.Errorf("invalid header hash: %w", err)
	}

	_, err = s.chainState.GetBlockByHash(headerHash)
	if err != nil {
		return "", NewRPCErrorWithData(ErrCodeBlockUnavailable, "block unavailable", headerHashStr)
	}

	// TODO: implement get service data from posterior state
	return encodeBlob([]byte{}), nil
}

func (s *RPCService) ServiceValue(headerHashStr string, serviceID uint64, key string) (string, error) {
	headerHash, err := decodeHash(headerHashStr)
	if err != nil {
		return "", fmt.Errorf("invalid header hash: %w", err)
	}

	_, err = s.chainState.GetBlockByHash(headerHash)
	if err != nil {
		return "", NewRPCErrorWithData(ErrCodeBlockUnavailable, "block unavailable", headerHashStr)
	}
	_, err = decodeBlob(key)
	if err != nil {
		return "", fmt.Errorf("invalid key blob: %w", err)
	}

	// TODO: implement get service value from posterior state
	return encodeBlob([]byte{}), nil
}

func (s *RPCService) ServicePreimage(headerHashStr string, serviceID uint64, hashStr string) (string, error) {
	headerHash, err := decodeHash(headerHashStr)
	if err != nil {
		return "", fmt.Errorf("invalid header hash: %w", err)
	}

	_, err = s.chainState.GetBlockByHash(headerHash)
	if err != nil {
		return "", NewRPCErrorWithData(ErrCodeBlockUnavailable, "block unavailable", headerHashStr)
	}
	_, err = decodeHash(hashStr)
	if err != nil {
		return "", fmt.Errorf("invalid preimage hash: %w", err)
	}

	// TODO: implement get service preimage from posterior state
	return encodeBlob([]byte{}), nil
}

func (s *RPCService) ServiceRequest(headerHashStr string, serviceID uint64, hashStr string, length uint32) (interface{}, error) {
	headerHash, err := decodeHash(headerHashStr)
	if err != nil {
		return nil, fmt.Errorf("invalid header hash: %w", err)
	}

	_, err = s.chainState.GetBlockByHash(headerHash)
	if err != nil {
		return nil, NewRPCErrorWithData(ErrCodeBlockUnavailable, "block unavailable", headerHashStr)
	}
	_, err = decodeHash(hashStr)
	if err != nil {
		return nil, fmt.Errorf("invalid request hash: %w", err)
	}

	// TODO: implement get service request from posterior state
	return nil, nil
}

func (s *RPCService) SubmitPreimage(requesterID uint64, preimage string) error {
	_, err := decodeBlob(preimage)
	if err != nil {
		return fmt.Errorf("invalid preimage blob: %w", err)
	}

	// TODO: implement submit service preimage
	return nil
}

func (s *RPCService) ListServices(headerHashStr string) ([]uint64, error) {
	logger.Debug("RPC ListServices called")
	headerHash, err := decodeHash(headerHashStr)
	if err != nil {
		return nil, fmt.Errorf("invalid header hash: %w", err)
	}

	_, err = s.chainState.GetBlockByHash(headerHash)
	logger.Debug("RPC ListServices: checked block existence")
	if err != nil {
		return nil, NewRPCErrorWithData(ErrCodeBlockUnavailable, "block unavailable", headerHashStr)
	}
	logger.Debug("RPC ListServices: block found")
	// TODO: implement get service IDs from posterior state
	return []uint64{}, nil
}

func (s *RPCService) SyncState() (*SyncState, error) {
	// TODO: implement sync state
	return &SyncState{
		NumPeers: 0,           // TODO: get actual number of peers
		Status:   "Completed", // TODO: get actual sync status
	}, nil
}

func (s *RPCService) WorkReport(hashStr string) (string, error) {
	_, err := decodeHash(hashStr)
	if err != nil {
		return "", fmt.Errorf("invalid header hash: %w", err)
	}

	// TODO: implement work report retrieval
	return encodeBlob([]byte{}), NewRPCErrorWithData(ErrCodeWorkReportUnavailable, "work report not implemented", nil)
}

func (s *RPCService) SubmitWorkPackage(core uint32, packageBlob string, extrinsic []string) error {
	_, err := decodeBlob(packageBlob)
	if err != nil {
		return fmt.Errorf("invalid package blob: %w", err)
	}

	for i, ext := range extrinsic {
		_, err := decodeBlob(ext)
		if err != nil {
			return fmt.Errorf("invalid extrinsic blob at index %d: %w", i, err)
		}
	}

	// TODO: implement work package submission
	return nil
}

func (s *RPCService) SubmitWorkPackageBundle(core uint32, bundle string) error {
	_, err := decodeBlob(bundle)
	if err != nil {
		return fmt.Errorf("invalid bundle blob: %w", err)
	}

	// TODO: implement work package bundle submission
	return nil
}

func (s *RPCService) WorkPackageStatus(headerHashStr string, hashStr string, anchorStr string) (interface{}, error) {
	_, err := decodeHash(headerHashStr)
	if err != nil {
		return nil, fmt.Errorf("invalid header hash: %w", err)
	}

	_, err = decodeHash(hashStr)
	if err != nil {
		return nil, fmt.Errorf("invalid work package hash: %w", err)
	}

	_, err = decodeBlob(anchorStr)
	if err != nil {
		return nil, fmt.Errorf("invalid anchor blob: %w", err)
	}

	// TODO: implement work package status retrieval
	return map[string]interface{}{
		"Failed": "Work-package status retrieval not implemented",
	}, nil
}

func (s *RPCService) FetchWorkPackageSegments(wpHashStr string, indices []uint32) ([]string, error) {
	_, err := decodeHash(wpHashStr)
	if err != nil {
		return nil, fmt.Errorf("invalid work package hash: %w", err)
	}

	// TODO: implement work package segment retrieval
	return make([]string, len(indices)), nil
}

func (s *RPCService) FetchSegments(segmentHashStr string, indices []uint32) ([]string, error) {
	_, err := decodeHash(segmentHashStr)
	if err != nil {
		return nil, fmt.Errorf("invalid segment hash: %w", err)
	}

	// TODO: implement segment retrieval
	return make([]string, len(indices)), nil
}

func encodeHash(hash [32]byte) string {
	return base64.StdEncoding.EncodeToString(hash[:])
}

func decodeHash(encoded string) (types.HeaderHash, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return types.HeaderHash{}, fmt.Errorf("invalid base64: %w", err)
	}
	if len(data) != 32 {
		return types.HeaderHash{}, fmt.Errorf("invalid hash length: expected 32 bytes, got %d bytes", len(data))
	}
	var hash types.HeaderHash
	copy(hash[:], data)
	return hash, nil
}

func encodeBlob(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func decodeBlob(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}

type BlockDescriptor struct {
	HeaderHash string `json:"header_hash"`
	Slot       uint64 `json:"slot"`
}

type SyncState struct {
	NumPeers int    `json:"num_peers"`
	Status   string `json:"status"`
}

type ChainParameters struct {
	V1 ChainParametersV1 `json:"V1"`
}

type ChainParametersV1 struct {
	DepositPerItem        uint64 `json:"deposit_per_item"`         // B_I
	DepositPerByte        uint64 `json:"deposit_per_byte"`         // B_L
	DepositPerAccount     uint64 `json:"deposit_per_account"`      // B_S
	CoreCount             uint32 `json:"core_count"`               // C
	MinTurnaroundPeriod   uint32 `json:"min_turnaround_period"`    // D
	EpochPeriod           uint32 `json:"epoch_period"`             // E
	MaxAccumulateGas      uint64 `json:"max_accumulate_gas"`       // G_A
	MaxIsAuthorizedGas    uint64 `json:"max_is_authorized_gas"`    // G_I
	MaxRefineGas          uint64 `json:"max_refine_gas"`           // G_R
	BlockGasLimit         uint64 `json:"block_gas_limit"`          // G_T
	RecentBlockCount      uint32 `json:"recent_block_count"`       // H
	MaxWorkItems          uint32 `json:"max_work_items"`           // I
	MaxDependencies       uint32 `json:"max_dependencies"`         // J
	MaxTicketsPerBlock    uint32 `json:"max_tickets_per_block"`    // K
	MaxLookupAnchorAge    uint32 `json:"max_lookup_anchor_age"`    // L
	TicketsAttemptsNumber uint32 `json:"tickets_attempts_number"`  // N
	AuthWindow            uint32 `json:"auth_window"`              // O
	SlotPeriodSec         uint32 `json:"slot_period_sec"`          // P
	AuthQueueLen          uint32 `json:"auth_queue_len"`           // Q
	RotationPeriod        uint32 `json:"rotation_period"`          // R
	MaxExtrinsics         uint32 `json:"max_extrinsics"`           // T
	AvailabilityTimeout   uint32 `json:"availability_timeout"`     // U
	ValCount              uint32 `json:"val_count"`                // V
	MaxAuthorizerCodeSize uint32 `json:"max_authorizer_code_size"` // W_A
	MaxInput              uint32 `json:"max_input"`                // W_B
	MaxServiceCodeSize    uint32 `json:"max_service_code_size"`    // W_C
	BasicPieceLen         uint32 `json:"basic_piece_len"`          // W_E
	MaxImports            uint32 `json:"max_imports"`              // W_M
	SegmentPieceCount     uint32 `json:"segment_piece_count"`      // W_P
	MaxReportElectiveData uint32 `json:"max_report_elective_data"` // W_R
	TransferMemoSize      uint32 `json:"transfer_memo_size"`       // W_T
	MaxExports            uint32 `json:"max_exports"`              // W_X
	EpochTailStart        uint32 `json:"epoch_tail_start"`         // Y
}
