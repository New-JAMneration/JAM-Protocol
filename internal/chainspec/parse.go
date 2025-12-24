package chainspec

import (
	"bytes"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func (cs *ChainSpec) GenesisHeaderBytes() ([]byte, error) {
	if cs == nil {
		return nil, fmt.Errorf("chainspec: nil")
	}
	return parseHexString(cs.GenesisHeader)
}

func (cs *ChainSpec) ProtocolParametersBytes() ([]byte, error) {
	s := strings.TrimSpace(cs.ProtocolParamsHex)
	if s == "" {
		return nil, fmt.Errorf("chainspec missing protocol_parameters")
	}
	return parseHexString(s)
}

func (cs *ChainSpec) GenesisStateKeyVals() (types.StateKeyVals, error) {
	if cs == nil {
		return nil, fmt.Errorf("chainspec: nil")
	}
	if cs.GenesisState == nil {
		return nil, fmt.Errorf("chainspec: genesis_state is nil")
	}

	out := make(types.StateKeyVals, 0, len(cs.GenesisState))
	for k, v := range cs.GenesisState {
		kb, err := parseHexString(k)
		if err != nil {
			return nil, fmt.Errorf("chainspec: genesis_state key %q: %w", k, err)
		}
		if len(kb) != 31 {
			return nil, fmt.Errorf("chainspec: genesis_state key must be 31 bytes (62 hex chars), got %d bytes for key %q", len(kb), k)
		}

		vb, err := parseHexString(v)
		if err != nil {
			return nil, fmt.Errorf("chainspec: genesis_state[%q] value: %w", k, err)
		}

		var sk types.StateKey
		copy(sk[:], kb)

		out = append(out, types.StateKeyVal{
			Key:   sk,
			Value: types.ByteSequence(vb),
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return bytes.Compare(out[i].Key[:], out[j].Key[:]) < 0
	})

	return out, nil
}

func (cs *ChainSpec) ParseProtocolParameters() (types.ProtocolParameters, error) {
	var pp types.ProtocolParameters

	ppBytes, err := cs.ProtocolParametersBytes()
	if err != nil {
		return types.ProtocolParameters{}, err
	}
	dec := types.NewDecoder()
	if err := dec.Decode(ppBytes, &pp); err != nil {
		return types.ProtocolParameters{}, fmt.Errorf("decode protocol parameters: %w", err)
	}
	return pp, nil
}

func normalizeHex(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "0x")
	return s
}

func parseHexString(s string) ([]byte, error) {
	s = normalizeHex(s)
	if s == "" {
		return nil, nil
	}
	if len(s)%2 != 0 {
		return nil, fmt.Errorf("invalid hex length (must be even), got %d", len(s))
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("invalid hex: %w", err)
	}
	return b, nil
}

var bootnodeBase32 = base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567").WithPadding(base32.NoPadding)

// ParseBootnode parses "<name>@<ip>:<port>".
// - name must be 53 chars, starts with 'e', rest is base32(pubkey) without padding.
// - ip must be a literal IPv4 or IPv6 address (IPv6 may be in [] or not).
func ParseBootnode(s string) (*Bootnode, error) {
	orig := strings.TrimSpace(s)
	if orig == "" {
		return nil, fmt.Errorf("bootnode: empty")
	}

	parts := strings.Split(orig, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("bootnode: must be <name>@<ip>:<port>, got %q", orig)
	}
	name := parts[0]
	addr := parts[1]

	if len(name) != 53 || !strings.HasPrefix(name, "e") {
		return nil, fmt.Errorf("bootnode: name must be 53 chars starting with 'e', got %q", name)
	}

	pubB32 := name[1:]
	pubBytes, err := bootnodeBase32.DecodeString(pubB32)
	if err != nil {
		return nil, fmt.Errorf("bootnode: invalid base32 pubkey in name %q: %w", name, err)
	}
	if len(pubBytes) != 32 {
		return nil, fmt.Errorf("bootnode: decoded pubkey must be 32 bytes, got %d", len(pubBytes))
	}
	var pk [32]byte
	copy(pk[:], pubBytes)

	host, portStr, err := splitHostPortLoose(addr)
	if err != nil {
		return nil, fmt.Errorf("bootnode: invalid addr %q: %w", addr, err)
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return nil, fmt.Errorf("bootnode: host must be an IP literal, got %q", host)
	}

	portU64, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil || portU64 == 0 {
		return nil, fmt.Errorf("bootnode: invalid port %q", portStr)
	}

	return &Bootnode{
		Original: orig,
		Name:     name,
		PubKey:   pk,
		IP:       ip,
		Port:     uint16(portU64),
	}, nil
}

// splitHostPortLoose accepts:
// - IPv4: "1.2.3.4:1234"
// - IPv6: "[::1]:1234" or "::1:1234" (loose form; split at last ':')
func splitHostPortLoose(addr string) (host string, port string, err error) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return "", "", fmt.Errorf("empty addr")
	}

	// Try strict parser first (handles [ipv6]:port and ipv4:port)
	h, p, e := net.SplitHostPort(addr)
	if e == nil {
		return h, p, nil
	}

	// Loose fallback: split on last ':'
	i := strings.LastIndex(addr, ":")
	if i <= 0 || i == len(addr)-1 {
		return "", "", fmt.Errorf("missing port")
	}
	host = addr[:i]
	port = addr[i+1:]
	host = strings.TrimPrefix(host, "[")
	host = strings.TrimSuffix(host, "]")
	return host, port, nil
}
