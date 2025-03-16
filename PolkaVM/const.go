package PolkaVM

// Memory related constants
const (
	ZA = 2       // PVM dynamic address alignment factor.
	ZI = 1 << 24 // standard PVM program initialization input data size.
	ZP = 1 << 12 // the PVM memory page size.
	ZZ = 1 << 16 // standard PVM program initialization zone size
)

// I.4.4 Constants
var (
	UnreferencedPreimageTimeslots = 28800 // D
)
