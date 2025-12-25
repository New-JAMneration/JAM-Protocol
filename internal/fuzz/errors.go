package fuzz

import "errors"

var (
	ErrNotImpl            = errors.New("not implemented")
	ErrInvalidVersion     = errors.New("invalid version string")
	ErrInvalidMessageType = errors.New("invalid message type")
	ErrInvalidPayloadType = errors.New("invalid payload type")
)
