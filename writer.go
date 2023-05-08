package datagram

import (
	"bytes"
	"encoding/binary"
)

// A Writer provides methods to write a UDP payload.
type Writer struct {
	buffer *bytes.Buffer
}

// Remaining returns the number of bytes that can be written into the payload.
func (w *Writer) Remaining() int {
	return w.buffer.Cap() - w.buffer.Len()
}

// WriteUint16 writes the argument as two bytes into the payload.
func (w *Writer) WriteUint16(v uint16) error {
	if w.buffer == nil {
		return ErrClosedWriter
	}
	if w.buffer.Cap() < w.buffer.Len()+2 {
		return ErrOverflow
	}
	return binary.Write(w.buffer, binary.BigEndian, v)
}

// WriteUint64 writes the argument as 8 bytes into the payload.
func (w *Writer) WriteUint64(v uint64) error {
	if w.buffer == nil {
		return ErrClosedWriter
	}
	if w.buffer.Cap() < w.buffer.Len()+8 {
		return ErrOverflow
	}
	return binary.Write(w.buffer, binary.BigEndian, v)
}

// WriteInt64 writes the argument as 8 bytes into the payload.
func (w *Writer) WriteInt64(v int64) error {
	if w.buffer == nil {
		return ErrClosedWriter
	}
	if w.buffer.Cap() < w.buffer.Len()+8 {
		return ErrOverflow
	}
	return binary.Write(w.buffer, binary.BigEndian, v)
}

// WriteFloat64 writes the argument as 8 bytes into the payload.
func (w *Writer) WriteFloat64(v float64) error {
	if w.buffer == nil {
		return ErrClosedWriter
	}
	if w.buffer.Cap() < w.buffer.Len()+8 {
		return ErrOverflow
	}
	return binary.Write(w.buffer, binary.BigEndian, v)
}

// Write the byte slice to the payload, preceded by a two byte length field.
func (w *Writer) Write(v []byte) (err error) {
	if w.buffer == nil {
		return ErrClosedWriter
	}
	if w.buffer.Cap() < len(v)+2 {
		return ErrOverflow
	}
	length := uint16(len(v))
	if err = binary.Write(w.buffer, binary.BigEndian, &length); err != nil {
		return
	}
	w.buffer.Write(v)
	return
}
