package rpc

import (
	"encoding/json"

	"github.com/New-JAMneration/JAM-Protocol/internal/eventbus"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// ChainReader provides read-only access to chain state.
type ChainReader interface {
	GetLatestBlock() types.Block
	GetFinalizedBlocks() []types.Block
	GetBlockByHash(headerHash types.HeaderHash) (types.Block, error)
	GetGenesisBlock() types.Block
}

// EventPublisher publishes events to subscribers.
type EventPublisher interface {
	Publish(event eventbus.Event)
}

// EventSubscriber manages event subscriptions.
type EventSubscriber interface {
	Subscribe(eventType eventbus.EventType, bufferSize int) (uint64, <-chan eventbus.Event)
	Unsubscribe(eventType eventbus.EventType, subID uint64)
}

type JSONRPCRequest struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string           `json:"method"`
	Params  *json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Result  interface{}      `json:"result,omitempty"`
	Error   *RPCError        `json:"error,omitempty"`
}

// MarshalJSON ensures JSON-RPC 2.0 compliance: sucess responses always
// include "result" (even if null), error responses omit it.
func (r *JSONRPCResponse) MarshalJSON() ([]byte, error) {
	if r.Error != nil {
		// Error response: include "error", omit "result"
		type alias struct {
			JSONRPC string           `json:"jsonrpc"`
			ID      *json.RawMessage `json:"id,omitempty"`
			Error   *RPCError        `json:"error"`
		}
		return json.Marshal(&alias{
			JSONRPC: r.JSONRPC,
			ID:      r.ID,
			Error:   r.Error,
		})
	}
	// Success response: always include "result" (null if nil)
	type alias struct {
		JSONRPC string           `json:"jsonrpc"`
		ID      *json.RawMessage `json:"id,omitempty"`
		Result  interface{}      `json:"result"`
	}
	return json.Marshal(&alias{
		JSONRPC: r.JSONRPC,
		ID:      r.ID,
		Result:  r.Result, // will be null in JSON if nil
	})
}

type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

const (
	ErrCodeParseError     = -32700
	ErrCodeInvalidRequest = -32600
	ErrCodeMethodNotFound = -32601
	ErrCodeInvalidParams  = -32602
	ErrCodeInternalError  = -32603
)

const (
	ErrCodeBlockUnavailable      = 1
	ErrCodeWorkReportUnavailable = 2
	ErrCodeDASegmentUnavailable  = 3
	ErrCodeOther                 = 0
)

func NewErrorResponse(id *json.RawMessage, code int, message string, data interface{}) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

func NewSuccessResponse(id *json.RawMessage, result interface{}) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

func NewRPCErrorWithData(code int, message string, data interface{}) *RPCError {
	return &RPCError{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

func (e *RPCError) Error() string {
	return e.Message
}

type JSONRPCNotification struct {
	JSONRPC string             `json:"jsonrpc"`
	Method  string             `json:"method"`
	Params  NotificationParams `json:"params"`
}

type NotificationParams struct {
	Subscription uint64      `json:"subscription"`
	Result       interface{} `json:"-"`
	Error        string      `json:"-"`
}

// MarshalJSON ensures JIP-2 comliance: notification params contain either
// "result" or "error", but not both.
func (p NotificationParams) MarshalJSON() ([]byte, error) {
	if p.Error != "" {
		type alias struct {
			Subscription uint64 `json:"subscription"`
			Error        string `json:"error"`
		}
		return json.Marshal(&alias{
			Subscription: p.Subscription,
			Error:        p.Error,
		})
	}
	type alias struct {
		Subscription uint64      `json:"subscription"`
		Result       interface{} `json:"result"`
	}
	return json.Marshal(&alias{
		Subscription: p.Subscription,
		Result:       p.Result, // will be null in JSON if nil
	})
}

func (p *NotificationParams) UnmarshalJSON(data []byte) error {
	type alias struct {
		Subscription uint64           `json:"subscription"`
		Result       *json.RawMessage `json:"result,omitempty"`
		Error        string           `json:"error,omitempty"`
	}
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	p.Subscription = a.Subscription
	p.Error = a.Error
	if a.Result != nil {
		var v interface{}
		if err := json.Unmarshal(*a.Result, &v); err != nil {
			return err
		}
		p.Result = v
	}
	return nil
}
