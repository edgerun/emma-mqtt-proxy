package mqtt

import (
	"errors"
	"io"
)

var StreamStateError = errors.New("illegal state (packet has not been consumed, or Next has not been called)")

// Streamer is a lower level abstraction to read from an MQTT packet stream in a flexible way. The streamer is stateful
// and the caller needs to first call Next to advance the streamer to the next packet, and then call either
// WritePacketTo or DecodePacket to consume the next packet from the stream.
type Streamer interface {
	Next() (*PacketHeader, error)
	WritePacketTo(io.Writer) error
	DecodePacket() (Packet, error)
}

// Reader is a higher level abstraction to read from an MQTT packet stream. Read will block until the next MQTT packet
// has been fully read from the stream.
type Reader interface {
	Read() (Packet, error)
}

// Writer writes an MQTT packet to an underlying data stream.
type Writer interface {
	// Write the packet to the underlying data stream.
	Write(packet Packet) error
}

// The StreamReader wraps a Streamer as a Reader.
type StreamReader struct {
	s Streamer
}

func NewStreamReader(s Streamer) *StreamReader {
	return &StreamReader{s}
}

func (sr *StreamReader) Read() (Packet, error) {
	header, err := sr.s.Next()
	if err != nil {
		return nil, err
	}
	if header == nil {
		return nil, errors.New("header was nil")
	}
	return sr.s.DecodePacket()
}
