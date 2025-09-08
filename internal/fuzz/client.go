package fuzz

import (
	"net"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type FuzzClient struct {
	conn net.Conn
}

func NewFuzzClient(network, address string) (*FuzzClient, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}

	return NewFuzzClientInternal(conn), nil
}

func NewFuzzClientInternal(conn net.Conn) *FuzzClient {
	client := FuzzClient{
		conn: conn,
	}

	return &client
}

func (c *FuzzClient) Close() error {
	return c.conn.Close()
}

func (c *FuzzClient) makeRequest(req Message) (Message, error) {
	reqBytes, err := req.MarshalBinary()
	if err != nil {
		return Message{}, err
	}

	_, err = c.conn.Write(reqBytes)
	if err != nil {
		return Message{}, err
	}

	var resp Message

	_, err = resp.ReadFrom(c.conn)
	if err != nil {
		return Message{}, err
	}

	return resp, nil
}

func (c *FuzzClient) Handshake(peerInfo PeerInfo) (PeerInfo, error) {
	req := Message{
		Type:     MessageType_PeerInfo,
		PeerInfo: &peerInfo,
	}

	resp, err := c.makeRequest(req)
	if err != nil {
		return PeerInfo{}, nil
	}

	if resp.Type != MessageType_PeerInfo {
		return PeerInfo{}, ErrInvalidMessageType
	}

	return *resp.PeerInfo, nil
}

func (c *FuzzClient) ImportBlock(block types.Block) (types.StateRoot, error) {
	payload := ImportBlock(block)

	req := Message{
		Type:        MessageType_ImportBlock,
		ImportBlock: &payload,
	}

	resp, err := c.makeRequest(req)
	if err != nil {
		return types.StateRoot{}, err
	}

	if resp.Type != MessageType_StateRoot {
		return types.StateRoot{}, ErrInvalidMessageType
	}

	return types.StateRoot(*resp.StateRoot), nil
}

func (c *FuzzClient) SetState(header types.Header, state types.StateKeyVals) (types.StateRoot, error) {
	payload := SetState{
		Header: header,
		State:  state,
	}

	req := Message{
		Type:     MessageType_SetState,
		SetState: &payload,
	}

	resp, err := c.makeRequest(req)
	if err != nil {
		return types.StateRoot{}, err
	}

	if resp.Type != MessageType_StateRoot {
		return types.StateRoot{}, ErrInvalidMessageType
	}

	return types.StateRoot(*resp.StateRoot), nil
}

func (c *FuzzClient) GetState(hash types.HeaderHash) (types.StateKeyVals, error) {
	payload := GetState(hash)

	req := Message{
		Type:     MessageType_GetState,
		GetState: &payload,
	}

	resp, err := c.makeRequest(req)
	if err != nil {
		return types.StateKeyVals{}, err
	}

	if resp.Type != MessageType_State {
		return types.StateKeyVals{}, ErrInvalidMessageType
	}

	// TODO
	return types.StateKeyVals(*resp.State), nil
}
