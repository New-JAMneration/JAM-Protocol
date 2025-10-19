package quic

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/quic-go/quic-go"
)

type Connection struct {
	Conn quic.Connection
	Role PeerRole
	Addr net.Addr
}

func NewConnection(conn quic.Connection, role PeerRole, addr net.Addr) *Connection {
	return &Connection{
		Conn: conn,
		Role: role,
		Addr: addr,
	}
}

func Dial(ctx context.Context, addr string, tlsConfig *tls.Config, quicConfig *quic.Config, role PeerRole) (*Connection, error) {
	conn, err := quic.DialAddr(ctx, addr, tlsConfig, quicConfig)
	if err != nil {
		return nil, err
	}
	return &Connection{
		Conn: conn,
		Role: role,
		Addr: conn.RemoteAddr(),
	}, nil
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

type ConnectionManager struct {
	addrMap map[string]*Connection
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		addrMap: make(map[string]*Connection),
	}
}

func (cm *ConnectionManager) GetByAddr(addr string) (*Connection, bool) {
	conn, exists := cm.addrMap[addr]
	return conn, exists
}

func (cm *ConnectionManager) Update(f func(*ConnectionManager) (interface{}, error)) (interface{}, error) {
	return f(cm)
}
