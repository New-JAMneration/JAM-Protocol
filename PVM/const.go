package PVM

import "github.com/New-JAMneration/JAM-Protocol/logger"

// Memory related constants
const (
	ZA = 2       // PVM dynamic address alignment factor.
	ZI = 1 << 24 // standard PVM program initialization input data size.
	ZP = 1 << 12 // the PVM memory page size.
	ZZ = 1 << 16 // standard PVM program initialization zone size
)

var (
	// PVM logger instance
	pvmLogger = logger.GetLogger("pvm")

	// log print as hex or dec, default: dec
	instrLogFormat = "dec"

	// perInstruction, blockBased, default: perInstruction
	GasChargingMode = "perInstruction"
)
