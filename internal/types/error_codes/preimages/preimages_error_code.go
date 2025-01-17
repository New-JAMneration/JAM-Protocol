package types

import (
	. "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type PreimageErrorCode ErrorCode

const (
	PreimageUnneeded PreimageErrorCode = iota // 0
)
