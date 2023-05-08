package datagram

import (
	"math"
)

// MaxPayload is the maximum data size for regular UDP datagrams.
const MaxPayload uint16 = math.MaxUint16 - 8 /* UDP */ - 20 /* IPv4 */
