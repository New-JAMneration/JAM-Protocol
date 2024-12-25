package types

// need to transform to byte to concat with the message
const (
	JamAvailable    = iota // 0
	JamBeefy               // 1
	JamEntropy             // 2
	JamFallbackSeal        // 3
	JamGuarantee           // 4
	JamAnnounce            // 5
	JamTicketSeal          // 6
	JamAudit               // 7
	JamVaild               // 8
	JamInvalid             // 9
)
