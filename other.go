package datagram

import (
	"errors"
	"os"
)

// IsTimeout returns true if the network action timed out.
func IsTimeout(err error) bool {
	return err != nil && errors.Is(err, os.ErrDeadlineExceeded)
}
