package fuzz

import (
	"net"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// TODO
type FuzzClient struct {
	conn net.Conn
}

func NewFuzzClient(network, address string) (*FuzzClient, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}

	client := FuzzClient{
		conn: conn,
	}

	return &client, nil
}

func (c *FuzzClient) Close() error {
	return c.conn.Close()
}

func (c *FuzzClient) Handshake(peerInfo PeerInfo) (PeerInfo, error) {
	m := Message{
		Type:    MessageType_PeerInfo,
		payload: &peerInfo,
	}

	req, err := m.MarshalBinary()
	if err != nil {
		return PeerInfo{}, err
	}

	_, err = c.conn.Write(req)
	if err != nil {
		return PeerInfo{}, err
	}

	var resp Message
	_, err = resp.ReadFrom(c.conn)
	if err != nil {
		return PeerInfo{}, err
	}

	info, err := resp.PeerInfo()
	if err != nil {
		return PeerInfo{}, err
	}

	return *info, nil
}

func (c *FuzzClient) ImportBlock(block types.Block) (StateRoot, error) {
	// TODO
	return StateRoot{}, ErrNotImpl
}

func (c *FuzzClient) SetState(header types.Header, state State) (StateRoot, error) {
	// TODO
	return StateRoot{}, ErrNotImpl
}

func (c *FuzzClient) GetState(hash types.HeaderHash) (State, error) {
	// TODO
	return State{}, ErrNotImpl
}
