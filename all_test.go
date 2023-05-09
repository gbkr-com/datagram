package datagram

import (
	"context"
	"math"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gbkr-com/app"
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
	t.Skip()
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
			if app.IsDone(ctx) {
				break
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
	addr, _ := net.ResolveUDPAddr("udp", "localhost:"+strconv.Itoa(rcvaddr.Port))
	sender.Send(w, addr, 20*time.Millisecond)
	<-time.After(20 * time.Millisecond)
	cxl()
}

func TestMultiple(t *testing.T) {
	// t.Skip()
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
	//
	// Control.
	//
	timeout := 10 * time.Millisecond
	blocking := new(sync.WaitGroup)
	var received []int64
	//
	// Receiving.
	//
	blocking.Add(1)
	go func() {
		defer blocking.Done()
		for {
			reader, _, err := receiver.Receive(timeout)
			if IsTimeout(err) {
				continue
			}
			if err != nil {
				t.Error(err)
			}
			v, err := reader.ReadInt64()
			if err != nil {
				t.Error(err)
			}
			reader.Close()
			received = append(received, v)
			if v == 10 {
				break
			}
		}
	}()
	//
	// Sending.
	//
	go func() {
		addr, _ := net.ResolveUDPAddr("udp", "localhost:"+strconv.Itoa(rcvaddr.Port))
		for i := 0; i < 10; i++ {
			writer := sender.Writer()
			writer.WriteInt64(int64(i + 1))
			err := sender.Send(writer, addr, timeout)
			if err != nil {
				t.Error(err)
			}
		}
	}()
	blocking.Wait()
	if len(received) != 10 {
		t.Error()
	}
}
