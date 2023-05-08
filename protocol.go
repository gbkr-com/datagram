package datagram

// A Protocol defines how to communicate over UDP. The hash is used in the
// payload header to filter out 'stranger' UDP datagrams. A non-zero hash will
// cause the protocol to be written first into every sent payload and read first
// for every received payload.
//
// The hash can be made with, for example:
//
//	h := maphash.String(maphash.MakeSeed(), "my-protocol/v1")
//
// The payload is the maximum data size expected with the protocol. Note
// the constant MaxPayload in this package.
type Protocol struct {
	Hash    uint64
	Payload uint16
}

func protocolWrite(protocol *Protocol, writer *Writer) error {
	return writer.WriteUint64(protocol.Hash)
}

func protocolRead(protocol *Protocol, reader *Reader) (ok bool, err error) {
	var hash uint64
	hash, err = reader.ReadUint64()
	if err != nil {
		return
	}
	if hash != protocol.Hash {
		return
	}
	ok = true
	return
}
