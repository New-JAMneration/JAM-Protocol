package fuzz

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
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
			return nil, fmt.Errorf("failed to remove existing socket file: %v", err)
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
				log.Printf("warning: failed to remove socket file %s: %v", unixAddr.Name, err)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			conn, err := s.Listener.Accept()
			if err != nil {
				log.Printf("error while accepting connection: %v", err)
				continue
			}

			go s.serve(ctx, conn)
		}
	}
}

func (s *FuzzServer) serve(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			var req, resp Message
			_, err := req.ReadFrom(conn)
			if err == io.EOF {
				log.Printf("[fuzz-server] connection closed")
				return
			}

			if err != nil {
				log.Printf("[fuzz-server] error while reading requests, err: %v", err)
				return
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
				log.Printf("[fuzz-server] error processing request[%v]: %v", req.Type, err)
				continue
			}

			respBytes, err := resp.MarshalBinary()
			if err != nil {
				log.Printf("[fuzz-server] error marshaling response: %v", err)
				continue
			}

			_, err = conn.Write(respBytes)
			if err != nil {
				log.Printf("[fuzz-server] error writing response: %v", err)
				continue
			}
		}
	}
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
	stateRoot, err := s.Service.SetState(m.SetState.Header, m.SetState.State)
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
		log.Printf("[fuzz-server][GetState] error: %v", err)
	}

	payload := State(stateKeyVals)

	return Message{
		Type:  MessageType_State,
		State: &payload,
	}, nil
}
