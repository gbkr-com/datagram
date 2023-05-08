package datagram

import (
	"bytes"
	"encoding/binary"
)

// A Reader provides methods to read a UDP payload.
type Reader struct {
	buffer   *bytes.Buffer
	endpoint *Endpoint
}

// ReadUint16 reads an uint64 from the payload.
func (r *Reader) ReadUint16() (v uint16, err error) {
	if r.buffer == nil {
		err = ErrClosedReader
		return
	}
	err = binary.Read(r.buffer, binary.BigEndian, &v)
	return
}

// ReadUint64 reads an uint64 from the payload.
func (r *Reader) ReadUint64() (v uint64, err error) {
	if r.buffer == nil {
		err = ErrClosedReader
		return
	}
	err = binary.Read(r.buffer, binary.BigEndian, &v)
	return
}

// ReadInt64 reads an int64 from the payload.
func (r *Reader) ReadInt64() (v int64, err error) {
	if r.buffer == nil {
		err = ErrClosedReader
		return
	}
	err = binary.Read(r.buffer, binary.BigEndian, &v)
	return
}

// ReadFloat64 reads a float64 from the payload.
func (r *Reader) ReadFloat64() (v float64, err error) {
	if r.buffer == nil {
		err = ErrClosedReader
		return
	}
	err = binary.Read(r.buffer, binary.BigEndian, &v)
	return
}

// Read a byte slice from the payload.
func (r *Reader) Read() (v []byte, err error) {
	if r.buffer == nil {
		err = ErrClosedReader
		return
	}
	var length uint16
	if err = binary.Read(r.buffer, binary.BigEndian, &length); err != nil {
		return
	}
	if length > MaxPayload {
		err = ErrOverflow
		return
	}
	v = make([]byte, length)
	_, err = r.buffer.Read(v)
	return
}

// Close the reader.
func (r *Reader) Close() error {
	if r.buffer == nil {
		return ErrClosedReader
	}
	r.endpoint.buffers.Recycle(r.buffer)
	r.buffer = nil
	return nil
}
