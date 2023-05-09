package datagram

import (
	"bytes"
	"net"
	"strconv"
	"time"

	"github.com/gbkr-com/app"
)

// An Endpoint for communication via UDP. Sending a UDP payload is done by:
//
//   - calling Writer() to get a reference to a writer
//   - using the writer methods to add to the payload
//   - calling Send() to send the payload to a given address.
//
// Likewise, a UDP payload is received by:
//
//   - calling Receive() to obtain a reader reference and the sending address
//   - using the reader methods to extract data from the payload
//   - calling Close() on the reader
//
// The end point minimises allocations by having a pool of buffers for
// sending and receiving.
type Endpoint struct {
	protocol *Protocol
	sequence uint64                   // Last written sequence number.
	conn     *net.UDPConn             // The underlying connection.
	zero     []byte                   // A zero filled payload.
	buffers  *app.Pool[*bytes.Buffer] // Pool of payload buffers, used by readers and writers.
	writers  *app.Pool[*Writer]       // Pool of writers.
}

// A Connection is the connection between this end point and a remote UDP address.
type Connection struct {
	remote *net.UDPAddr
}

// NewEndpoint returns a UDP end point that is connected to the network.
// The pool specifies how many buffers to keep for recycing.
//
// This function will panic in a number of circumstances:
//   - if the protocol is nil.
//   - if the given protocol payload is zero or greater than MaxPayload.
//   - if the protocol requires verification but the payload size is less than 8 bytes.
//   - if the port is negative.
//   - if the pool size is less than one.
func NewEndpoint(protocol *Protocol, port, pool int) (*Endpoint, error) {
	if protocol == nil {
		panic("protocol")
	}
	if protocol.Payload == 0 || protocol.Payload > MaxPayload {
		panic("payload")
	}
	if protocol.Hash > 0 && protocol.Payload < 8 {
		panic("hash")
	}
	if port < 0 {
		panic("port")
	}
	if pool < 1 {
		panic("pool")
	}
	//
	// Make the net.UDPConn.
	//
	var hostport string
	if port > 0 {
		hostport = ":" + strconv.Itoa(port)
	}
	addr, err := net.ResolveUDPAddr("udp", hostport)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	//
	// Return the end point.
	//
	e := &Endpoint{
		protocol: protocol,
		conn:     conn,
		zero:     make([]byte, protocol.Payload),
		buffers: app.NewPool(
			pool,
			app.WithPoolFactory(
				func() *bytes.Buffer {
					buffer := new(bytes.Buffer)
					buffer.Grow(int(protocol.Payload))
					return buffer
				},
			),
			app.WithPoolReset(
				func(b *bytes.Buffer) {
					b.Reset()
				},
			),
			app.WithPoolDiscard[*bytes.Buffer](),
		),
		writers: app.NewPool(
			pool,
			app.WithPoolFactory(func() *Writer { return &Writer{} }),
			app.WithPoolReset(func(w *Writer) { w.buffer = nil }),
			app.WithPoolDiscard[*Writer](),
		),
	}
	return e, nil
}

// LocalAddress returns the address of this end point.
func (e *Endpoint) LocalAddress() *net.UDPAddr {
	return e.conn.LocalAddr().(*net.UDPAddr)
}

// LastSequence returns the last written sequence number.
func (e *Endpoint) LastSequence() uint64 {
	return e.sequence
}

// SetSequence sets the last written sequence number.
func (e *Endpoint) SetSequence(seq uint64) {
	e.sequence = seq
}

func (e *Endpoint) incr() {
	e.sequence++
}

// Writer returns a new writer.
func (e *Endpoint) Writer() *Writer {
	w := e.writers.Next()
	w.buffer = e.buffers.Next()
	if e.protocol.Hash > 0 {
		protocolWrite(e.protocol, w)
	}
	if e.protocol.Sequenced {
		sequenceWrite(e, w)
	}
	return w
}

// Send the UDP payload in the writer from this end point. The writer should not
// be used again after this call.
func (e *Endpoint) Send(writer *Writer, address *net.UDPAddr, timeout time.Duration) (err error) {
	if timeout > 0 {
		if err := e.conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
			return err
		}
	}
	_, err = e.conn.WriteToUDP(writer.buffer.Bytes(), address)
	if err != nil {
		return
	}
	e.buffers.Recycle(writer.buffer)
	e.writers.Recycle(writer)
	return
}

// Receive a UDP payload. The returned reader is used to extract items from
// the payload. That reader must be closed after use.
// The returned reader may be nil: this happens when there is an error and also
// when the incoming UDP datagram does not match the protocol.
func (e *Endpoint) Receive(timeout time.Duration) (reader *Reader, addr *net.UDPAddr, seq uint64, err error) {
	if timeout > 0 {
		if err = e.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			return
		}
	}
	//
	// Get a buffer and fill it, then use the underlying byte slice for the
	// ReadFromUDP operation.
	//
	buffer := e.buffers.Next()
	buffer.Write(e.zero)
	bx := buffer.Bytes()
	var n int
	if n, addr, err = e.conn.ReadFromUDP(bx); err != nil {
		return
	}
	//
	// Although the byte slice has been manipulated outside of the buffer we can
	// still get the buffer back to normal by truncating to the number of bytes
	// put into the slice by the ReadFromUDP.
	//
	buffer.Truncate(n)
	reader = &Reader{
		buffer:   buffer,
		endpoint: e,
	}
	if e.protocol.Hash > 0 {
		var ok bool
		ok, err = protocolRead(e.protocol, reader)
		if err != nil || !ok {
			reader = nil
			addr = nil
			return
		}
	}
	if e.protocol.Sequenced {
		seq, err = sequenceRead(e, reader)
	}
	return
}

// Close this end point.
func (e *Endpoint) Close() error {
	return e.conn.Close()
}
