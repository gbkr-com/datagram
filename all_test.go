package datagram

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testprotocol = Protocol{
	Hash:    0, // maphash.String(maphash.MakeSeed(), "testing"),
	Payload: 256,
}

func BenchmarkBuffering(b *testing.B) {
	//
	// Create the end point.
	//
	sender, _ := NewEndpoint(&testprotocol, 0, 8)
	defer sender.Close()
	//
	// Benchmark.
	//
	for i := 0; i < b.N; i++ {
		buff := sender.buffers.Next()
		sender.buffers.Recycle(buff)
	}
}

func BenchmarkWriting(b *testing.B) {
	//
	// Create the end points.
	//
	sender, _ := NewEndpoint(&testprotocol, 0, 8)
	defer sender.Close()
	data := make([]byte, 64)
	//
	// Benchmark.
	//
	for i := 0; i < b.N; i++ {
		w := sender.Writer()
		w.Write(data)
		sender.buffers.Recycle(w.buffer) // mimics what happens after the send.
		sender.writers.Recycle(w)
	}
}

func BenchmarkWritingAndSending(b *testing.B) {
	//
	// Create the end points.
	//
	sender, _ := NewEndpoint(&testprotocol, 0, 8)
	defer sender.Close()
	receiver, _ := NewEndpoint(&testprotocol, 0, 8)
	defer receiver.Close()
	remote := receiver.LocalAddress()
	data := make([]byte, 64)
	//
	// Benchmark.
	//
	for i := 0; i < b.N; i++ {
		w := sender.Writer()
		w.Write(data)
		sender.Send(w, remote, 0)
	}
}

func TestPayloadSize(t *testing.T) {
	//
	// Create the end point.
	//
	sender, err := NewEndpoint(&testprotocol, 0, 8)
	assert.Nil(t, err)
	defer sender.Close()
	//
	// Get a writer then write until it breaks.
	//
	w := sender.Writer()
	assert.Equal(t, 256, w.Remaining())
	err = w.Write(make([]byte, 250))
	assert.Nil(t, err)
	assert.Equal(t, 4, w.Remaining())
	err = w.WriteFloat64(0)
	assert.NotNil(t, err)
}

func TestSendAndReceive(t *testing.T) {
	//
	// Create a sender and a receiver.
	//
	receiver, err := NewEndpoint(&testprotocol, 0, 8)
	assert.Nil(t, err)
	defer receiver.Close()
	rcvaddr := receiver.LocalAddress()
	sender, err := NewEndpoint(&testprotocol, 0, 8)
	assert.Nil(t, err)
	defer sender.Close()
	sndaddr := sender.LocalAddress()
	//
	// Receive in a goroutine that can be cancelled.
	//
	ctx, cxl := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-ctx.Done():
				break
			default:
			}
			reader, address, err := receiver.Receive(20 * time.Millisecond)
			if err != nil {
				if IsTimeout(err) {
					return
				}
				t.Error(err)
			}
			if address.Port != sndaddr.Port {
				t.Error()
			}
			b, err := reader.Read()
			if err != nil {
				t.Error(err)
			}
			if string(b) != "hello world" {
				t.Error()
			}
			f, err := reader.ReadFloat64()
			if err != nil {
				t.Error(err)
			}
			if f != math.Pi {
				t.Error()
			}
			err = reader.Close()
			assert.Nil(t, err)
			return
		}
	}()
	w := sender.Writer()
	w.Write([]byte("hello world"))
	w.WriteFloat64(math.Pi)
	sender.Send(w, rcvaddr, 20*time.Millisecond)
	<-time.After(20 * time.Millisecond)
	cxl()
}
