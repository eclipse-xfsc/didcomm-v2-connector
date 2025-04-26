package intErr

import "errors"

var (
	ErrNoPingResponseRequested = errors.New("no ping response requested")
	ErrUnknownMessageType      = errors.New("unknown message type")
	ErrNotImplemented          = errors.New("not implemented")
	ErrUnpackingMessage        = errors.New("can not unpacking received message")
)
