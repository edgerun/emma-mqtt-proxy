package mqtt

import (
	"errors"
	"io"
)

var StreamStateError = errors.New("illegal state (packet has not been consumed, or Next has not been called)")

// Streamer is a lower level abstraction to read from an MQTT packet stream in a flexible way. The streamer is stateful
// and the caller needs to first call Next to advance the streamer to the next packet.
type Streamer interface {
	// Advances the stream and reads the next header. The packet is not yet read from the underlying stream.
	Next() (*PacketHeader, error)

	// Reads the current packet from the underlying stream, decodes, and returns it.
	ReadPacket() (Packet, error)
}

// Reader is a higher level abstraction to read from an MQTT packet stream. ReadPacket will block until the next MQTT packet
// has been fully read from the stream.
type Reader interface {
	ReadPacket() (Packet, error)
}

type ReaderFrom interface {
	ReadPacketFrom(Reader) error
}

type WriterTo interface {
	WritePacketTo(Writer) error
}

// Writer writes an MQTT packet to an underlying data stream.
type Writer interface {
	// WritePacket the packet to the underlying data stream.
	WritePacket(packet Packet) error
}

type PacketSink interface {
	ReaderFrom
	Writer
}

// The StreamReader wraps a Streamer as a Reader that calls Next on the Streamer implicitly when getting ReadPacket.
type StreamReader struct {
	s Streamer
}

func NewStreamReader(s Streamer) *StreamReader {
	return &StreamReader{s}
}

func (sr *StreamReader) ReadPacket() (Packet, error) {
	return ReadNext(sr.s)
}

func ReadNext(s Streamer) (Packet, error) {
	header, err := s.Next()
	if err != nil {
		return nil, err
	}
	if header == nil {
		return nil, errors.New("header was nil")
	}
	return s.ReadPacket()
}

// Copy copies the current packet from the source Reader to the destination Writer. If possible, it uses ReaderFrom and
// WriterTo interfaces to give the underlying implementations an opportunity to optimize the copying process. Otherwise
// it simply reads (and potentially decodes) the next packet and writes it to the writer.
func Copy(src Reader, dst Writer) error {
	if wt, ok := dst.(WriterTo); ok {
		err := wt.WritePacketTo(dst)
		return err
	}
	if rf, ok := dst.(ReaderFrom); ok {
		err := rf.ReadPacketFrom(src)
		return err
	}

	packet, err := src.ReadPacket()
	if err != nil {
		return err
	}
	return dst.WritePacket(packet)
}

type Channel interface {
	Streamer
	PacketSink
}

type codecChannel struct {
	*DecodingStreamer
	*Encoder
}

func NewChannel(rw io.ReadWriter) Channel {
	return &codecChannel{
		NewDecodingStreamer(rw),
		NewEncoder(rw),
	}
}
