package telemetry

import "fmt"

// NodeInfo is the connection-initial message JIP-3 requires before any
// event. Unlike events it has NO Timestamp / Discriminator header — the
// first byte on the wire is ProtocolVersion. Putting Timestamp at the
// start would shift every subsequent field by 8 bytes and JamTART would
// fail to decode.
//
// Wire field order (set by Encode):
//
//  1. Protocol version            (u8 = 0)
//  2. JAM Parameters              (canonical bytes from GP fetch host call)
//  3. Genesis header hash         ([32]byte)
//  4. Peer ID                     ([32]byte Ed25519 public key)
//  5. Peer IPv6 + port            ([16]byte ++ u16 LE)
//  6. Node flags                  (u32 LE; bit 0 = PVM recompiler)
//  7. Implementation name         (String<32>)
//  8. Implementation version      (String<32>)
//  9. Gray Paper version          (String<16>)
//  10. Freeform info              (String<512>; may be empty)
type NodeInfo struct {
	// JAMParameters is opaque pass-through bytes for now.
	// TODO(jip3-canonical-jam-params): replace with a typed encoder once
	// the canonical encoding source is confirmed (#775 Q3).
	JAMParameters []byte

	GenesisHash [32]byte
	PeerID      [32]byte // Ed25519 public key
	PeerIPv6    [16]byte
	PeerPort    uint16
	NodeFlags   uint32

	ImplName     string // String<32>
	ImplVersion  string // String<32>
	GrayPaperVer string // String<16>
	FreeformInfo string // String<512>; may be empty
}

// NodeInfoProtocolVersion is JIP-3's protocol version byte (0 for the
// current spec revision).
const NodeInfoProtocolVersion uint8 = 0

// Per-field String<N> caps; names match JIP-3 column headings.
const (
	maxImplName     uint32 = 32
	maxImplVersion  uint32 = 32
	maxGrayPaperVer uint32 = 16
	maxFreeformInfo uint32 = 512
)

// NodeFlagPVMRecompiler is bit 0 of NodeFlags. Set when using the PVM
// recompiler, clear for the interpreter. All other bits must be 0.
const NodeFlagPVMRecompiler uint32 = 1 << 0

// Encode produces the wire bytes for NodeInfo (NOT including the outer
// 4-byte length prefix that the framing layer adds). Errors on
// undefined NodeFlags bits, oversized String<N>, or invalid UTF-8.
func (n NodeInfo) Encode() ([]byte, error) {
	// Reject undefined NodeFlags up front so a local mistake doesn't
	// reach the aggregator's parser.
	if n.NodeFlags&^NodeFlagPVMRecompiler != 0 {
		return nil, fmt.Errorf(
			"NodeInfo.NodeFlags: undefined bits set: 0x%08x (defined mask 0x%08x)",
			n.NodeFlags, NodeFlagPVMRecompiler,
		)
	}

	// Validate Strings first so length-cap / UTF-8 errors surface before
	// any concatenation cost.
	implName, err := EncodeString(n.ImplName, maxImplName)
	if err != nil {
		return nil, fmt.Errorf("NodeInfo.ImplName: %w", err)
	}
	implVer, err := EncodeString(n.ImplVersion, maxImplVersion)
	if err != nil {
		return nil, fmt.Errorf("NodeInfo.ImplVersion: %w", err)
	}
	gpVer, err := EncodeString(n.GrayPaperVer, maxGrayPaperVer)
	if err != nil {
		return nil, fmt.Errorf("NodeInfo.GrayPaperVer: %w", err)
	}
	freeform, err := EncodeString(n.FreeformInfo, maxFreeformInfo)
	if err != nil {
		return nil, fmt.Errorf("NodeInfo.FreeformInfo: %w", err)
	}

	out := make([]byte, 0,
		1+ // protocol version
			len(n.JAMParameters)+
			32+ // genesis hash
			32+ // peer ID
			16+2+ // IPv6 + port
			4+ // node flags
			len(implName)+len(implVer)+len(gpVer)+len(freeform))

	out = append(out, NodeInfoProtocolVersion)
	out = append(out, n.JAMParameters...)
	out = append(out, n.GenesisHash[:]...)
	out = append(out, n.PeerID[:]...)
	out = append(out, n.PeerIPv6[:]...)
	out = append(out, EncodeU16(n.PeerPort)...)
	out = append(out, EncodeU32(n.NodeFlags)...)
	out = append(out, implName...)
	out = append(out, implVer...)
	out = append(out, gpVer...)
	out = append(out, freeform...)
	return out, nil
}
