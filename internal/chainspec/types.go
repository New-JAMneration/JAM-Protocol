package chainspec

import "net"

// ChainSpec matches JIP-4 JSON schema.
type ChainSpec struct {
	ID                string            `json:"id"`
	Bootnodes         []string          `json:"bootnodes,omitempty"`
	GenesisHeader     string            `json:"genesis_header"`
	GenesisState      map[string]string `json:"genesis_state"`
	ProtocolParamsHex string            `json:"protocol_parameters,omitempty"`
}

// Bootnode is a parsed bootnode entry <name>@<ip>:<port>.
type Bootnode struct {
	Original string   // original string
	Name     string   // 53 chars: 'e' + base32(pubkey)
	PubKey   [32]byte // decoded Ed25519 public key
	IP       net.IP   // IPv4 or IPv6 address
	Port     uint16   // port number
}
