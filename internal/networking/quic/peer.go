package quic

import (
	"context"
	"crypto/ed25519"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"sync"

	networkcert "github.com/New-JAMneration/JAM-Protocol/internal/networking/cert"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/quic-go/quic-go"
)

// HeadInfo represents information about a block header
type HeadInfo struct {
	Hash     types.HeaderHash `json:"hash"`     // Data32 -> types.HeaderHash
	Timeslot types.TimeSlot   `json:"timeslot"` // TimeslotIndex -> types.TimeSlot
}

type PeerRole string

const (
	Validator PeerRole = "validator"
	Builder   PeerRole = "builder"
)

type PeerConfig struct {
	Role          PeerRole
	Addr          net.Addr
	GenesisHeader types.HeaderHash
	PublicKey     ed25519.PublicKey
	UPHandler     UPHandler
	CEHandler     CEHandler
}

type StreamHandlerFunc func(ctx context.Context, stream *Stream, peerKey ed25519.PublicKey) error

type Peer struct {
	Ed25519Key     ed25519.PublicKey
	ValidatorIndex *uint16
	Listener       *Listener
	tlsConfig      *tls.Config
	quicConfig     *quic.Config
	connManager    *ConnectionManager
	peerSet        *PeerSet
	ctx            context.Context
	cancel         context.CancelFunc
	handlerMu      sync.RWMutex
	handlers       map[byte]StreamHandlerFunc
	CEHandler      CEHandler
	UPHandler      UPHandler
	// Sync-related fields
	Best      *HeadInfo `json:"best"`      // Optional best block header
	Finalized HeadInfo  `json:"finalized"` // Finalized block header
	ID        string    `json:"id"`        // Peer identifier
}

func NewPeer(config PeerConfig) (*Peer, error) {

	isBuilder := config.Role == Builder

	tlsConfig, err := NewTLSConfig(false, isBuilder)

	if err != nil {
		return nil, err
	}

	quicConfig := NewQuicConfig()

	listener, err := NewListener(config.Addr.String(), isBuilder, NewTLSConfig, quicConfig)

	if err != nil {
		return nil, err
	}

	return &Peer{
		Ed25519Key:  config.PublicKey,
		Listener:    listener,
		tlsConfig:   tlsConfig,
		quicConfig:  quicConfig,
		connManager: NewConnectionManager(),
		peerSet:     NewPeerSet(),
		handlers:    make(map[byte]StreamHandlerFunc),
		CEHandler:   config.CEHandler,
		UPHandler:   config.UPHandler,

		// TODO: use a better ID
		ID: config.Addr.String() + string(config.PublicKey),
	}, nil
}

func (p *Peer) Connect(addr net.Addr, role PeerRole) (*Connection, error) {
	if existingConn, ok := p.connManager.GetByAddr(addr.String()); ok {
		return existingConn, nil
	}

	conn, err := Dial(context.TODO(), addr.String(), p.tlsConfig, p.quicConfig, role)
	if err != nil {
		return nil, err
	}
	p.connManager.Add(addr.String(), conn)
	return conn, nil
}

func (p *Peer) RegisterHandler(kind byte, h StreamHandlerFunc) {
	p.handlerMu.Lock()
	defer p.handlerMu.Unlock()
	p.handlers[kind] = h
}

func (p *Peer) Start(ctx context.Context) error {
	if p.Listener == nil {
		return fmt.Errorf("listener is nil")
	}
	if p.cancel != nil {
		return nil
	}

	p.ctx, p.cancel = context.WithCancel(ctx)
	go p.acceptLoop()
	return nil
}

func (p *Peer) acceptLoop() {
	for {
		select {
		case <-p.ctx.Done():
			return
		default:
		}

		qconn, err := p.Listener.Accept(p.ctx)
		if err != nil {
			if p.ctx.Err() != nil {
				return
			}
			log.Println("accept connection error:", err)
			continue
		}

		peerKey, err := extractPeerKey(qconn)
		if err != nil {
			log.Println("extract peer key error:", err)
			_ = qconn.CloseWithError(0, "invalid peer certificate")
			continue
		}

		go p.handleConnection(qconn, peerKey)
	}
}

func (p *Peer) handleConnection(qconn quic.Connection, peerKey ed25519.PublicKey) {
	conn := NewConnection(qconn, Validator, qconn.RemoteAddr())
	p.connManager.Add(conn.Addr.String(), conn)

	remote := &Peer{Ed25519Key: peerKey}
	if err := p.peerSet.Add(remote, conn.Addr.String()); err != nil {
		log.Println("peer set add error:", err)
	}

	for {
		stream, err := conn.AcceptStream(p.ctx)
		if err != nil {
			if p.ctx != nil && p.ctx.Err() != nil {
				return
			}
			return
		}
		go p.dispatchStream(&Stream{Stream: stream}, peerKey)
	}
}

func (p *Peer) dispatchStream(stream *Stream, peerKey ed25519.PublicKey) {
	kind, err := stream.ReadStreamKind()
	if err != nil {
		log.Println("read stream kind error:", err)
		_ = stream.Close()
		return
	}

	p.handlerMu.RLock()
	handler, ok := p.handlers[kind]
	p.handlerMu.RUnlock()
	if !ok {
		log.Printf("no handler for stream kind: %d", kind)
		_ = stream.Close()
		return
	}

	if err := handler(p.ctx, stream, peerKey); err != nil {
		log.Println("stream handler error:", err)
		_ = stream.Close()
	}
}

func (p *Peer) Broadcast(kind string, message interface{}) {
	msg, err := p.UPHandler.EncodeMessage(kind, message)

	if err != nil {
		log.Println("error encoding message:", err)
		return
	}

	kindByte := byte(0)
	if len(kind) == 1 {
		kindByte = kind[0]
	}

	for _, conn := range p.connManager.All() {
		stream, err := conn.OpenStreamSync(context.Background())

		if err != nil {
			log.Println("error opening stream:", err)
			continue
		}

		wrapped := &Stream{Stream: stream}
		if err := wrapped.WriteStreamKind(kindByte); err != nil {
			log.Println("error writing stream kind:", err)
			_ = wrapped.Close()
			continue
		}

		if _, err := wrapped.Write(msg); err != nil {
			log.Println("error writing to stream:", err)
			_ = wrapped.Close()
			continue
		}
		_ = wrapped.Close()
	}
}

// SetTLSInsecureSkipVerify sets the InsecureSkipVerify field of the TLS config
// This is useful for testing purposes to skip certificate verification
func (p *Peer) SetTLSInsecureSkipVerify(skip bool) {
	if p.tlsConfig != nil {
		p.tlsConfig.InsecureSkipVerify = skip
	}
}

func extractPeerKey(conn quic.Connection) (ed25519.PublicKey, error) {
	peerCerts := conn.ConnectionState().TLS.PeerCertificates
	if len(peerCerts) == 0 {
		return nil, fmt.Errorf("missing peer certificate")
	}

	if err := networkcert.ValidateX509Certificate(peerCerts[0]); err != nil {
		return nil, err
	}

	pub, ok := peerCerts[0].PublicKey.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("peer key is not ed25519")
	}
	return pub, nil
}
