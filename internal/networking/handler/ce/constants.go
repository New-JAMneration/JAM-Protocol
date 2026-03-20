package ce

// CE protocol wire-format sizes (bytes). Use these instead of magic numbers.
const (
	// Common: hash and integer sizes
	HashSize = 32 // OpaqueHash, HeaderHash, ErasureRoot, etc.
	U8Size   = 1  // uint8 little-endian
	U16Size  = 2  // uint16 little-endian
	U32Size  = 4  // uint32 little-endian

	// CE128 Block Request: HeaderHash (32) + Direction (1) + MaxBlocks (4)
	CE128MinRequestSize = HashSize + 1 + U32Size

	// CE129 State Request: HeaderHash (32) + KeyStart (31) + KeyEnd (31) + MaxSize (4)
	StateKeySize     = 31
	CE129RequestSize = HashSize + StateKeySize + StateKeySize + U32Size

	// CE131 Safrole Ticket Distribution: EpochIndex (4) + Attempt (1) + RingVRF Proof (784)
	CE131ProofSize   = 784                          // RingVRF proof size in bytes
	CE131PayloadSize = U32Size + 1 + CE131ProofSize // 789

	// CE133 Work Package Submission: minimum first message = CoreIndex (u16)
	CE133MinFirstMessageSize = U16Size

	// CE137 EC Shard Request: ErasureRoot (32) + ShardIndex (u16)
	CE137RequestSize = HashSize + U16Size
	// CE138 Audit Shard Request: ErasureRoot (32) + ShardIndex (u16)
	CE138RequestSize = HashSize + U16Size

	// CE139/CE140 Segment Shard Request: ErasureRoot (32) + ShardIndex (u16) + SegmentIndicesLen (u16)
	CE139140MinRequestSize = HashSize + U16Size + U16Size
	// Max segment indices per request (2*W_M, W_M=3072 per JAMNP)
	MaxSegmentIndicesCount = 6144
	SegmentIndexSize       = U16Size

	// Justification encoding: 1-byte discriminator + 32-byte hash per entry
	JustificationDiscriminatorSize = 1
	JustificationHashEntrySize     = JustificationDiscriminatorSize + HashSize
)
