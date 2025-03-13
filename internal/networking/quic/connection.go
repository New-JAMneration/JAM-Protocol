package quic

import (
	"context"
	"crypto/tls"

	"github.com/quic-go/quic-go"
)

type Connection struct {
	Conn quic.Connection
}

func Dial(ctx context.Context, addr string, tlsConfig *tls.Config, quicConfig *quic.Config) (*Connection, error) {
	conn, err := quic.DialAddr(ctx, addr, tlsConfig, quicConfig)
	if err != nil {
		return nil, err
	}
	return &Connection{Conn: conn}, nil
}

func (c *Connection) AcceptStream(ctx context.Context) (quic.Stream, error) {
	return c.Conn.AcceptStream(ctx)
}

func (c *Connection) OpenStream(ctx context.Context) (quic.Stream, error) {
	return c.Conn.OpenStreamSync(ctx)
}

func (c *Connection) OpenStreamSync(ctx context.Context) (quic.Stream, error) {
	return c.Conn.OpenStreamSync(ctx)
}

func (c *Connection) OpenStreamAsync() (quic.Stream, error) {
	return c.Conn.OpenStream()
}

func (c *Connection) Close() error {
	return c.Conn.CloseWithError(0, "closing")
}
