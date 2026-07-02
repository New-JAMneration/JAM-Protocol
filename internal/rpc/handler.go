package rpc

import (
	"encoding/json"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/eventbus"
	"github.com/New-JAMneration/JAM-Protocol/logger"
)

type Handler struct {
	service    *RPCService
	subManager *SubscriptionManager
}

func NewHandler(service *RPCService, subManager *SubscriptionManager) *Handler {
	return &Handler{
		service:    service,
		subManager: subManager,
	}
}

func (h *Handler) HandleMessage(message []byte) []byte {
	logger.Debug(fmt.Sprintf("Parsing JSON-RPC request: %s", string(message)))

	var req JSONRPCRequest

	if err := json.Unmarshal(message, &req); err != nil {
		logger.Error(fmt.Sprintf("JSON parse error: %v", err))
		return h.marshalResponse(NewErrorResponse(nil, ErrCodeParseError, "Parse error", nil))
	}

	if req.JSONRPC != "2.0" {
		logger.Error(fmt.Sprintf("Invalid JSON-RPC version: %s", req.JSONRPC))
		return h.marshalResponse(NewErrorResponse(req.ID, ErrCodeInvalidRequest, "Invalid Request", nil))
	}

	if req.Method == "" {
		logger.Error("Missing method in JSON-RPC request")
		return h.marshalResponse(NewErrorResponse(req.ID, ErrCodeInvalidRequest, "Invalid Request", nil))
	}

	logger.Debug(fmt.Sprintf("Calling method: %s", req.Method))

	resp := h.routeMethod(&req)

	return h.marshalResponse(resp)
}

func (h *Handler) routeMethod(req *JSONRPCRequest) *JSONRPCResponse {
	switch req.Method {
	case "ping":
		return h.handlePing(req)
	case "parameters":
		return h.handleParameters(req)
	case "bestBlock":
		return h.handleBestBlock(req)
	case "finalizedBlock":
		return h.handleFinalizedBlock(req)
	case "parent":
		return h.handleParent(req)
	case "stateRoot":
		return h.handleStateRoot(req)
	case "beefyRoot":
		return h.handleBeefyRoot(req)
	case "statistics":
		return h.handleStatistics(req)
	case "serviceData":
		return h.handleServiceData(req)
	case "serviceValue":
		return h.handleServiceValue(req)
	case "servicePreimage":
		return h.handleServicePreimage(req)
	case "serviceRequest":
		return h.handleServiceRequest(req)
	case "submitPreimage":
		return h.handleSubmitPreimage(req)
	case "listServices":
		return h.handleListServices(req)
	case "syncState":
		return h.handleSyncState(req)

	case "workReport":
		return h.handleWorkReport(req)
	case "submitWorkPackage":
		return h.handleSubmitWorkPackage(req)
	case "submitWorkPackageBundle":
		return h.handleSubmitWorkPackageBundle(req)
	case "workPackageStatus":
		return h.handleWorkPackageStatus(req)
	case "fetchWorkPackageSegments":
		return h.handleFetchWorkPackageSegments(req)
	case "fetchSegments":
		return h.handleFetchSegments(req)

	case "subscribeBestBlock":
		return h.handleSubscribeBestBlock(req)
	case "subscribeFinalizedBlock":
		return h.handleSubscribeFinalizedBlock(req)
	case "subscribeStatistics":
		return h.handleSubscribeStatistics(req)
	case "subscribeServiceData":
		return h.handleSubscribeServiceData(req)
	case "subscribeServiceValue":
		return h.handleSubscribeServiceValue(req)
	case "subscribeServicePreimage":
		return h.handleSubscribeServicePreimage(req)
	case "subscribeServiceRequest":
		return h.handleSubscribeServiceRequest(req)
	case "subscribeWorkPackageStatus":
		return h.handleSubscribeWorkPackageStatus(req)
	case "subscribeSyncStatus":
		return h.handleSubscribeSyncStatus(req)
	case "unsubscribe",
		"unsubscribeBestBlock",
		"unsubscribeFinalizedBlock",
		"unsubscribeStatistics",
		"unsubscribeServiceData",
		"unsubscribeServiceValue",
		"unsubscribeServicePreimage",
		"unsubscribeServiceRequest",
		"unsubscribeWorkPackageStatus",
		"unsubscribeSyncStatus":
		return h.handleUnsubscribe(req)

	default:
		logger.Warn(fmt.Sprintf("Method not found: %s", req.Method))
		return NewErrorResponse(req.ID, ErrCodeMethodNotFound, "Method not found", nil)
	}
}

func (h *Handler) handlePing(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Ping -> Pong")
	return NewSuccessResponse(req.ID, "pong")
}

func (h *Handler) handleParameters(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Getting parameters")

	result, err := h.service.Parameters()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get parameters: %v", err))
		return NewErrorResponse(req.ID, ErrCodeInternalError, "Internal error", nil)
	}

	return NewSuccessResponse(req.ID, result)
}

func (h *Handler) handleBestBlock(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Getting best block")

	result, err := h.service.BestBlock()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get best block: %v", err))
		return NewErrorResponse(req.ID, ErrCodeInternalError, "Internal error", nil)
	}

	return NewSuccessResponse(req.ID, result)
}

func (h *Handler) handleFinalizedBlock(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Getting finalized block")

	result, err := h.service.FinalizedBlock()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get finalized block: %v", err))
		return NewErrorResponse(req.ID, ErrCodeInternalError, "Internal error", nil)
	}

	return NewSuccessResponse(req.ID, result)
}

func (h *Handler) handleParent(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Getting parent block")

	if req.Params == nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}
	var params []string
	if err := json.Unmarshal(*req.Params, &params); err != nil || len(params) != 1 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}

	result, err := h.service.Parent(params[0])
	if err != nil {
		if rpcErr, ok := err.(*RPCError); ok {
			return NewErrorResponse(req.ID, rpcErr.Code, rpcErr.Message, rpcErr.Data)
		}
		return NewErrorResponse(req.ID, ErrCodeInternalError, err.Error(), nil)
	}

	return NewSuccessResponse(req.ID, result)
}

func (h *Handler) handleStateRoot(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Getting state root")

	if req.Params == nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}
	var params []string
	if err := json.Unmarshal(*req.Params, &params); err != nil || len(params) != 1 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}

	result, err := h.service.StateRoot(params[0])
	if err != nil {
		if rpcErr, ok := err.(*RPCError); ok {
			return NewErrorResponse(req.ID, rpcErr.Code, rpcErr.Message, rpcErr.Data)
		}
		return NewErrorResponse(req.ID, ErrCodeInternalError, err.Error(), nil)
	}

	return NewSuccessResponse(req.ID, result)
}

func (h *Handler) handleBeefyRoot(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Getting beefy root")

	if req.Params == nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}
	var params []string
	if err := json.Unmarshal(*req.Params, &params); err != nil || len(params) != 1 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}

	result, err := h.service.BeefyRoot(params[0])
	if err != nil {
		if rpcErr, ok := err.(*RPCError); ok {
			return NewErrorResponse(req.ID, rpcErr.Code, rpcErr.Message, rpcErr.Data)
		}
		return NewErrorResponse(req.ID, ErrCodeInternalError, err.Error(), nil)
	}

	return NewSuccessResponse(req.ID, result)
}

func (h *Handler) handleStatistics(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Getting statistics")

	if req.Params == nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}
	var params []string
	if err := json.Unmarshal(*req.Params, &params); err != nil || len(params) != 1 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}

	result, err := h.service.Statistics(params[0])
	if err != nil {
		if rpcErr, ok := err.(*RPCError); ok {
			return NewErrorResponse(req.ID, rpcErr.Code, rpcErr.Message, rpcErr.Data)
		}
		return NewErrorResponse(req.ID, ErrCodeInternalError, err.Error(), nil)
	}

	return NewSuccessResponse(req.ID, result)
}

func (h *Handler) handleServiceData(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Getting service data")

	if req.Params == nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}
	var params []interface{}
	if err := json.Unmarshal(*req.Params, &params); err != nil || len(params) != 2 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}

	headerHash, ok1 := params[0].(string)
	serviceID, ok2 := params[1].(float64)
	if !ok1 || !ok2 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params types", nil)
	}

	result, err := h.service.ServiceData(headerHash, uint64(serviceID))
	if err != nil {
		if rpcErr, ok := err.(*RPCError); ok {
			return NewErrorResponse(req.ID, rpcErr.Code, rpcErr.Message, rpcErr.Data)
		}
		return NewErrorResponse(req.ID, ErrCodeInternalError, err.Error(), nil)
	}

	if result == "" {
		return NewSuccessResponse(req.ID, nil)
	}

	return NewSuccessResponse(req.ID, result)
}

func (h *Handler) handleServiceValue(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Getting service value")

	if req.Params == nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}
	var params []interface{}
	if err := json.Unmarshal(*req.Params, &params); err != nil || len(params) != 3 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}

	headerHash, ok1 := params[0].(string)
	serviceID, ok2 := params[1].(float64)
	key, ok3 := params[2].(string)
	if !ok1 || !ok2 || !ok3 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params types", nil)
	}

	result, err := h.service.ServiceValue(headerHash, uint64(serviceID), key)
	if err != nil {
		if rpcErr, ok := err.(*RPCError); ok {
			return NewErrorResponse(req.ID, rpcErr.Code, rpcErr.Message, rpcErr.Data)
		}
		return NewErrorResponse(req.ID, ErrCodeInternalError, err.Error(), nil)
	}

	if result == "" {
		return NewSuccessResponse(req.ID, nil)
	}

	return NewSuccessResponse(req.ID, result)
}

func (h *Handler) handleServicePreimage(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Getting service preimage")

	if req.Params == nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}
	var params []interface{}
	if err := json.Unmarshal(*req.Params, &params); err != nil || len(params) != 3 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}

	headerHash, ok1 := params[0].(string)
	serviceID, ok2 := params[1].(float64)
	hash, ok3 := params[2].(string)
	if !ok1 || !ok2 || !ok3 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params types", nil)
	}

	result, err := h.service.ServicePreimage(headerHash, uint64(serviceID), hash)
	if err != nil {
		if rpcErr, ok := err.(*RPCError); ok {
			return NewErrorResponse(req.ID, rpcErr.Code, rpcErr.Message, rpcErr.Data)
		}
		return NewErrorResponse(req.ID, ErrCodeInternalError, err.Error(), nil)
	}

	if result == "" {
		return NewSuccessResponse(req.ID, nil)
	}

	return NewSuccessResponse(req.ID, result)
}

func (h *Handler) handleServiceRequest(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Getting service request")

	if req.Params == nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}
	var params []interface{}
	if err := json.Unmarshal(*req.Params, &params); err != nil || len(params) != 4 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}

	headerHash, ok1 := params[0].(string)
	serviceID, ok2 := params[1].(float64)
	hash, ok3 := params[2].(string)
	length, ok4 := params[3].(float64)
	if !ok1 || !ok2 || !ok3 || !ok4 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params types", nil)
	}

	result, err := h.service.ServiceRequest(headerHash, uint64(serviceID), hash, uint32(length))
	if err != nil {
		if rpcErr, ok := err.(*RPCError); ok {
			return NewErrorResponse(req.ID, rpcErr.Code, rpcErr.Message, rpcErr.Data)
		}
		return NewErrorResponse(req.ID, ErrCodeInternalError, err.Error(), nil)
	}

	return NewSuccessResponse(req.ID, result)
}

func (h *Handler) handleSubmitPreimage(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Submitting service preimage")

	if req.Params == nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}
	var params []interface{}
	if err := json.Unmarshal(*req.Params, &params); err != nil || len(params) != 2 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}
	requesterID, ok1 := params[0].(float64)
	preimageData, ok2 := params[1].(string)
	if !ok1 || !ok2 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params types", nil)
	}

	err := h.service.SubmitPreimage(uint64(requesterID), preimageData)
	if err != nil {
		if rpcErr, ok := err.(*RPCError); ok {
			return NewErrorResponse(req.ID, rpcErr.Code, rpcErr.Message, rpcErr.Data)
		}
		return NewErrorResponse(req.ID, ErrCodeInternalError, err.Error(), nil)
	}

	return NewSuccessResponse(req.ID, nil)
}

func (h *Handler) handleListServices(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Listing services")

	if req.Params == nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}
	var params []string
	if err := json.Unmarshal(*req.Params, &params); err != nil || len(params) != 1 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}

	result, err := h.service.ListServices(params[0])
	if err != nil {
		if rpcErr, ok := err.(*RPCError); ok {
			return NewErrorResponse(req.ID, rpcErr.Code, rpcErr.Message, rpcErr.Data)
		}
		return NewErrorResponse(req.ID, ErrCodeInternalError, err.Error(), nil)
	}

	return NewSuccessResponse(req.ID, result)
}

func (h *Handler) handleSyncState(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Getting sync state")

	result, err := h.service.SyncState()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get sync state: %v", err))
		return NewErrorResponse(req.ID, ErrCodeInternalError, "Internal error", nil)
	}

	return NewSuccessResponse(req.ID, result)
}

func (h *Handler) handleWorkReport(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Handling work report")

	if req.Params == nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}
	var params []string
	if err := json.Unmarshal(*req.Params, &params); err != nil || len(params) != 1 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}

	result, err := h.service.WorkReport(params[0])
	if err != nil {
		if rpcErr, ok := err.(*RPCError); ok {
			return NewErrorResponse(req.ID, rpcErr.Code, rpcErr.Message, rpcErr.Data)
		}
		return NewErrorResponse(req.ID, ErrCodeInternalError, err.Error(), nil)
	}

	return NewSuccessResponse(req.ID, result)
}

func (h *Handler) handleSubmitWorkPackage(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Submitting work package")

	if req.Params == nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}
	var params []interface{}
	if err := json.Unmarshal(*req.Params, &params); err != nil || len(params) != 3 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}

	core, ok1 := params[0].(float64)
	packageBlob, ok2 := params[1].(string)
	extrinsicRaw, ok3 := params[2].([]interface{})
	if !ok1 || !ok2 || !ok3 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params types", nil)
	}

	extrinsics := make([]string, len(extrinsicRaw))
	for i, ex := range extrinsicRaw {
		exStr, ok := ex.(string)
		if !ok {
			return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid extrinsic type", nil)
		}
		extrinsics[i] = exStr
	}

	err := h.service.SubmitWorkPackage(uint32(core), packageBlob, extrinsics)
	if err != nil {
		return NewErrorResponse(req.ID, ErrCodeInternalError, err.Error(), nil)
	}

	return NewSuccessResponse(req.ID, nil)
}

func (h *Handler) handleSubmitWorkPackageBundle(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Submitting work package bundle")

	if req.Params == nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}
	var params []interface{}
	if err := json.Unmarshal(*req.Params, &params); err != nil || len(params) != 2 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}

	core, ok1 := params[0].(float64)
	bundle, ok2 := params[1].(string)
	if !ok1 || !ok2 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params types", nil)
	}

	err := h.service.SubmitWorkPackageBundle(uint32(core), bundle)
	if err != nil {
		return NewErrorResponse(req.ID, ErrCodeInternalError, err.Error(), nil)
	}

	return NewSuccessResponse(req.ID, nil)
}

func (h *Handler) handleWorkPackageStatus(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Getting work package status")

	if req.Params == nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}
	var params []string
	if err := json.Unmarshal(*req.Params, &params); err != nil || len(params) != 3 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}

	result, err := h.service.WorkPackageStatus(params[0], params[1], params[2])
	if err != nil {
		return NewErrorResponse(req.ID, ErrCodeInternalError, err.Error(), nil)
	}

	return NewSuccessResponse(req.ID, result)
}

func (h *Handler) handleFetchWorkPackageSegments(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Fetching work package segments")

	if req.Params == nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}
	var params []interface{}
	if err := json.Unmarshal(*req.Params, &params); err != nil || len(params) != 2 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}

	wpHash, ok1 := params[0].(string)
	indicesRaw, ok2 := params[1].([]interface{})
	if !ok1 || !ok2 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params types", nil)
	}

	indices := make([]uint32, len(indicesRaw))
	for i, idx := range indicesRaw {
		idxFloat, ok := idx.(float64)
		if !ok {
			return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid index type", nil)
		}
		indices[i] = uint32(idxFloat)
	}

	result, err := h.service.FetchWorkPackageSegments(wpHash, indices)
	if err != nil {
		return NewErrorResponse(req.ID, ErrCodeInternalError, err.Error(), nil)
	}

	return NewSuccessResponse(req.ID, result)
}

func (h *Handler) handleFetchSegments(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Fetching segments")

	if req.Params == nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}
	var params []interface{}
	if err := json.Unmarshal(*req.Params, &params); err != nil || len(params) != 2 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}

	segmentRoot, ok1 := params[0].(string)
	indicesRaw, ok2 := params[1].([]interface{})
	if !ok1 || !ok2 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params types", nil)
	}

	indices := make([]uint32, len(indicesRaw))
	for i, idx := range indicesRaw {
		idxFloat, ok := idx.(float64)
		if !ok {
			return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid index type", nil)
		}
		indices[i] = uint32(idxFloat)
	}

	result, err := h.service.FetchSegments(segmentRoot, indices)
	if err != nil {
		return NewErrorResponse(req.ID, ErrCodeInternalError, err.Error(), nil)
	}

	return NewSuccessResponse(req.ID, result)
}

func (h *Handler) handleSubscribeBestBlock(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Subscribing to best block")

	subID, err := h.subManager.Subscribe(eventbus.EventNewBlock, "subscribeBestBlock")
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to subscribe: %v", err))
		return NewErrorResponse(req.ID, ErrCodeInternalError, "Failed to create subscription", nil)
	}

	return NewSuccessResponse(req.ID, subID)
}

func (h *Handler) handleSubscribeFinalizedBlock(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Subscribing to finalized block")

	subID, err := h.subManager.Subscribe(eventbus.EventFinalizedBlock, "subscribeFinalizedBlock")
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to subscribe: %v", err))
		return NewErrorResponse(req.ID, ErrCodeInternalError, "Failed to create subscription", nil)
	}

	return NewSuccessResponse(req.ID, subID)
}

func (h *Handler) handleSubscribeStatistics(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Subscribing to statistics")

	// Parse finalized boolean parameter per JIP-2: subscribeStatistics(finalized)
	if req.Params == nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}
	var params []interface{}
	if err := json.Unmarshal(*req.Params, &params); err != nil || len(params) != 1 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params: expected [finalized]", nil)
	}
	finalized, ok := params[0].(bool)
	if !ok {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params: finalized must be a boolean", nil)
	}

	// Use finalized flag to choose between best-chain and finalized-chain event triggers.
	eventType := eventbus.EventNewBlock
	if finalized {
		eventType = eventbus.EventFinalizedBlock
	}

	// Transform BlockEvent into ChainSubscriptionUpdate per JIP-2.
	// TODO: populate Value with actual statistics blob once implemented.
	transform := func(event eventbus.Event) interface{} {
		blockEvent, ok := event.Data.(eventbus.BlockEvent)
		if !ok {
			return event.Data
		}
		return eventbus.ChainSubscriptionUpdate{
			HeaderHash: blockEvent.HeaderHash,
			Slot:       blockEvent.Slot,
			Value:      nil,
		}
	}
	subID, err := h.subManager.Subscribe(eventType, "subscribeStatistics", transform)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to subscribe: %v", err))
		return NewErrorResponse(req.ID, ErrCodeInternalError, "Failed to create subscription", nil)
	}

	return NewSuccessResponse(req.ID, subID)
}

func (h *Handler) handleSubscribeServiceData(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Subscribing to service data")

	// JIP-2: subscribeServiceData(service_id, finalized)
	if req.Params == nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}
	var params []interface{}
	if err := json.Unmarshal(*req.Params, &params); err != nil || len(params) != 2 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params: expected [service_id, finalized]", nil)
	}

	// TODO: implement service data subscription logic
	return NewErrorResponse(req.ID, ErrCodeInternalError, "subscribeServiceData not implemented", nil)
}

func (h *Handler) handleSubscribeServiceValue(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Subscribing to service value")

	// JIP-2: subscribeServiceValue(service_id, key, finalized)
	if req.Params == nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}
	var params []interface{}
	if err := json.Unmarshal(*req.Params, &params); err != nil || len(params) != 3 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params: expected [service_id, key, finalized]", nil)
	}
	// TODO: implement service value subscription logic
	return NewErrorResponse(req.ID, ErrCodeInternalError, "subscribeServiceValue not implemented", nil)
}

func (h *Handler) handleSubscribeServicePreimage(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Subscribing to service preimage")

	// JIP-2: subscribeServicePreimage(service_id, hash, finalized)
	if req.Params == nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}
	var params []interface{}
	if err := json.Unmarshal(*req.Params, &params); err != nil || len(params) != 3 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params: expected [service_id, hash, finalized]", nil)
	}
	// TODO: implement service preimage subscription logic
	return NewErrorResponse(req.ID, ErrCodeInternalError, "subscribeServicePreimage not implemented", nil)
}

func (h *Handler) handleSubscribeServiceRequest(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Subscribing to service request")

	// JIP-2: subscribeServiceRequest(service_id, hash, length, finalized)
	if req.Params == nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}
	var params []interface{}
	if err := json.Unmarshal(*req.Params, &params); err != nil || len(params) != 4 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params: expected [service_id, hash, length, finalized]", nil)
	}
	// TODO: implement service request subscription logic
	return NewErrorResponse(req.ID, ErrCodeInternalError, "subscribeServiceRequest not implemented", nil)
}

func (h *Handler) handleSubscribeWorkPackageStatus(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Subscribing to work package status")

	// JIP-2: subscribeWorkPackageStatus(hash, anchor, finalized)
	if req.Params == nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}
	var params []interface{}
	if err := json.Unmarshal(*req.Params, &params); err != nil || len(params) != 3 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params: expected [hash, anchor, finalized]", nil)
	}
	// TODO: implement workpackage status subscription logic
	return NewErrorResponse(req.ID, ErrCodeInternalError, "subscribeWorkPackageStatus not implemented", nil)
}

func (h *Handler) handleSubscribeSyncStatus(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Subscribing to sync status")

	subID, err := h.subManager.Subscribe(eventbus.EventSyncStateChanged, "subscribeSyncStatus")
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to subscribe: %v", err))
		return NewErrorResponse(req.ID, ErrCodeInternalError, "Failed to create subscription", nil)
	}

	return NewSuccessResponse(req.ID, subID)
}

func (h *Handler) handleUnsubscribe(req *JSONRPCRequest) *JSONRPCResponse {
	logger.Debug("Unsubscribing")
	if req.Params == nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}

	var params []interface{}
	if err := json.Unmarshal(*req.Params, &params); err != nil || len(params) != 1 {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid params", nil)
	}

	subIDFloat, ok := params[0].(float64)
	if !ok {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid subscription ID type, expected number", nil)
	}
	subID := uint64(subIDFloat)

	if err := h.subManager.Unsubscribe(subID); err != nil {
		logger.Error(fmt.Sprintf("Failed to unsubscribe: %v", err))
		return NewErrorResponse(req.ID, ErrCodeInternalError, err.Error(), nil)
	}

	return NewSuccessResponse(req.ID, true)
}

func (h *Handler) marshalResponse(resp *JSONRPCResponse) []byte {
	bytes, err := json.Marshal(resp)
	if err != nil {
		// In case of marshal error, return a internal error response
		logger.Error(fmt.Sprintf("Failed to marshal JSON-RPC response: %v", err))
		return []byte(`{"jsonrpc":"2.0","error":{"code":-32603,"message":"Internal error"}}`)
	}
	logger.Debug(fmt.Sprintf("JSON-RPC response: %s", string(bytes)))
	return bytes
}
