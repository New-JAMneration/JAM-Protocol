package quic

import (
	"context"
	"crypto/ed25519"
	"crypto/tls"
	"log"
	"net"

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

type Peer struct {
	publicKey   ed25519.PublicKey
	Listener    *Listener
	tlsConfig   *tls.Config
	quicConfig  *quic.Config
	connManager *ConnectionManager
	CEHandler   CEHandler
	UPHandler   UPHandler
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
		publicKey:   config.PublicKey,
		Listener:    listener,
		tlsConfig:   tlsConfig,
		quicConfig:  quicConfig,
		connManager: NewConnectionManager(),
		CEHandler:   config.CEHandler,
		UPHandler:   config.UPHandler,

		// TODO: use a better ID
		ID: config.Addr.String() + string(config.PublicKey),
	}, nil
}

func (p *Peer) Connect(addr net.Addr, role Peer) (*Connection, error) {
	conn, err := p.connManager.Update(func(cm *ConnectionManager) (interface{}, error) {
		if existingConn, ok := cm.GetByAddr(addr.String()); ok {
			return existingConn, nil
		}

		quicConn, err := Dial(context.TODO(), addr.String(), p.tlsConfig, p.quicConfig, Validator)

		if err != nil {
			return nil, err
		}

		cm.addrMap[addr.String()] = quicConn
		return quicConn, nil
	})

	if err != nil {
		return nil, err
	}

	return conn.(*Connection), nil
}

func (p *Peer) Broadcast(kind string, message interface{}) {
	msg, err := p.UPHandler.EncodeMessage(kind, message)

	if err != nil {
		log.Println("error encoding message:", err)
		return
	}

	for _, conn := range p.connManager.addrMap {
		stream, err := conn.OpenStreamSync(context.Background())

		if err != nil {
			log.Println("error opening stream:", err)
			continue
		}

		if _, err := stream.Write(msg); err != nil {
			log.Println("error writing to stream:", err)
			stream.Close()
			continue
		}
	}
}

// SetTLSInsecureSkipVerify sets the InsecureSkipVerify field of the TLS config
// This is useful for testing purposes to skip certificate verification
func (p *Peer) SetTLSInsecureSkipVerify(skip bool) {
	if p.tlsConfig != nil {
		p.tlsConfig.InsecureSkipVerify = skip
	}
}
