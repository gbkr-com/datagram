package datagram

import (
	"errors"
)

// Errors for this package.
var (
	ErrOverflow     = errors.New("overflow")
	ErrClosedWriter = errors.New("closed writer")
	ErrClosedReader = errors.New("closed reader")
)
