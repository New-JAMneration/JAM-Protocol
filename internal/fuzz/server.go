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

// TODO
// blocks until terminated
func (s *FuzzServer) ListenAndServe(ctx context.Context) error {
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

			go func() {
				defer conn.Close()

				// TODO
			}()
		}
	}
}
