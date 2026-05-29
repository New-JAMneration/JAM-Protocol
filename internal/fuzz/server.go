package fuzz

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/logger"
)

// TODO
type FuzzServer struct {
	Listener net.Listener
	Service  FuzzService
}

func NewFuzzServer(network, address string) (*FuzzServer, error) {
	// For Unix sockets, remove the socket file if it exists
	if network == "unix" {
		if err := os.Remove(address); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to remove existing socket file: %w", err)
		}
	}

	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}

	stub := new(FuzzServiceStub)
	server := FuzzServer{
		Listener: listener,
		Service:  stub,
	}

	return &server, nil
}

// blocks until terminated
func (s *FuzzServer) ListenAndServe(ctx context.Context) error {
	defer func() {
		s.Listener.Close()
		// Clean up Unix socket file when server stops
		if unixAddr, ok := s.Listener.Addr().(*net.UnixAddr); ok {
			if err := os.Remove(unixAddr.Name); err != nil && !os.IsNotExist(err) {
				logger.Warnf("failed to remove socket file %s: %v", unixAddr.Name, err)
			}
		}
	}()

	go func() {
		<-ctx.Done()
		s.Listener.Close()
	}()

	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			logger.Warnf("[fuzz-server] error accepting connection: %v", err)
			continue
		}

		go s.serve(ctx, conn)
	}
}

func (s *FuzzServer) serve(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	for {
		if ctx.Err() != nil {
			return
		}
		if err := s.serveOneRequest(conn); err != nil {
			if err == io.EOF {
				logger.Debugf("[fuzz-server] connection closed")
				return
			}
			logger.Errorf("[fuzz-server] %v", err)
			return
		}
	}
}

func (s *FuzzServer) serveOneRequest(conn net.Conn) (err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("[fuzz-server] panic: %v", r)
			err = fmt.Errorf("panic handling request: %v", r)
		}
	}()

	var req, resp Message
	_, err = req.ReadFrom(conn)
	if err != nil {
		return err
	}

	switch req.Type {
	case MessageType_PeerInfo:
		resp, err = s.handlePeerInfo(req)
	case MessageType_ImportBlock:
		resp, err = s.handleImportBlock(req)
	case MessageType_SetState:
		resp, err = s.handleSetState(req)
	case MessageType_GetState:
		resp, err = s.handleGetState(req)
	default:
		err = ErrInvalidMessageType
	}

	if err != nil {
		return fmt.Errorf("error processing request[%v]: %w", req.Type, err)
	}

	respBytes, err := resp.MarshalBinary()
	if err != nil {
		return fmt.Errorf("error marshaling response: %w", err)
	}

	if _, err = conn.Write(respBytes); err != nil {
		return fmt.Errorf("error writing response: %w", err)
	}

	return nil
}

func (s *FuzzServer) handlePeerInfo(m Message) (Message, error) {
	peerInfo, err := s.Service.Handshake(*m.PeerInfo)
	if err != nil {
		return Message{}, err
	}

	return Message{
		Type:     MessageType_PeerInfo,
		PeerInfo: &peerInfo,
	}, nil
}

func (s *FuzzServer) handleImportBlock(m Message) (Message, error) {
	stateRoot, err := s.Service.ImportBlock(types.Block(*m.ImportBlock))
	if err != nil {
		// 1. runtime/system error → fatal → close connection
		if strings.Contains(err.Error(), "STF runtime error") {
			return Message{}, err
		}
		// 2. protocol error → return ErrorMessage
		return Message{
			Type: MessageType_ErrorMessage,
			Error: &ErrorMessage{
				Error: err.Error(),
			},
		}, nil
	}

	payload := StateRoot(stateRoot)

	return Message{
		Type:      MessageType_StateRoot,
		StateRoot: &payload,
	}, nil
}

func (s *FuzzServer) handleSetState(m Message) (Message, error) {
	stateRoot, err := s.Service.SetState(m.SetState.Header, m.SetState.State, m.SetState.Ancestry)
	if err != nil {
		return Message{}, err
	}

	payload := StateRoot(stateRoot)

	return Message{
		Type:      MessageType_StateRoot,
		StateRoot: &payload,
	}, nil
}

func (s *FuzzServer) handleGetState(m Message) (Message, error) {
	stateKeyVals, err := s.Service.GetState(types.HeaderHash(*m.GetState))
	if err != nil {
		logger.Errorf("[fuzz-server][GetState] error: %v", err)
	}

	payload := State(stateKeyVals)

	return Message{
		Type:  MessageType_State,
		State: &payload,
	}, nil
}
