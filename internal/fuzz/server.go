package fuzz

import (
	"context"
	"log"
	"net"
)

// TODO
type FuzzServer struct {
	Listener net.Listener
	Service  FuzzService
}

func NewFuzzServer(filename string) (*FuzzServer, error) {
	listener, err := net.Listen("unix", filename)
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

// TODO
// blocks until terminated
func (s *FuzzServer) ListenAndServe(ctx context.Context) error {
	defer s.Listener.Close()

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
			var m Message

			_, err := m.ReadFrom(conn)
			if err != nil {
				log.Printf("error while reading requests: %v\n", err)
				return
			}

			switch m.Type {
			case MessageType_PeerInfo:
				payload, err := m.PeerInfo()
				if err != nil {
					log.Printf("error reading request: %v\n", err)
					continue
				}

				info, err := s.Service.Handshake(*payload)
				if err != nil {
					log.Printf("error servicing handshake request: %v\n", err)
					continue
				}

				resp := Message{
					Type:    MessageType_PeerInfo,
					payload: &info,
				}
				respBytes, err := resp.MarshalBinary()
				if err != nil {
					log.Printf("error marshaling response: %v\n", err)
					continue
				}

				_, err = conn.Write(respBytes)
				if err != nil {
					log.Printf("error writing response: %v\n", err)
					continue
				}
			case MessageType_ImportBlock:
				// TODO
				log.Println("not implemented")
			case MessageType_SetState:
				// TODO
				log.Println("not implemented")
			case MessageType_GetState:
				// TODO
				log.Println("not implemented")
			}
		}
	}
}
