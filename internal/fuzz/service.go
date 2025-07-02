package fuzz

import "context"

type FuzzService interface {
	Handshake(PeerInfo) PeerInfo
	ImportBlock(ImportBlock) StateRoot
	SetState(SetState) StateRoot
	GetState(GetState) StateRoot
}

// TODO
type FuzzServer struct {
}

// TODO
func NewFuzzServer() *FuzzServer {
	return nil
}

// TODO
// blocks until terminated
func (s *FuzzServer) ListenAndServe(ctx context.Context) error {
	return nil
}
