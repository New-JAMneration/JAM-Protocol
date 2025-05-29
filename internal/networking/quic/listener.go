package quic

import (
	"context"
	"crypto/tls"
	"log"

	"github.com/quic-go/quic-go"
)

type Listener struct {
	Listener *quic.Listener
}

// NewListener QUIC listener, addr (e.g. "localhost:4242")
func NewListener(addr string, isBuilder bool, tlsConfigProvider func(isServer, isBuilder bool) (*tls.Config, error), quicConfig *quic.Config) (*Listener, error) {
	tlsConfig, err := tlsConfigProvider(true, isBuilder)
	if err != nil {
		return nil, err
	}
	listener, err := quic.ListenAddr(addr, tlsConfig, quicConfig)
	if err != nil {
		return nil, err
	}
	log.Printf("QUIC Listener started on %s", addr)
	return &Listener{Listener: listener}, nil
}

// Accept (blocking call)
func (l *Listener) Accept(ctx context.Context) (quic.Connection, error) {
	return l.Listener.Accept(ctx)
}

func (l *Listener) ListenAddress() string {
	return l.Listener.Addr().String()
}

func (l *Listener) Close() error {
	return l.Listener.Close()
}
