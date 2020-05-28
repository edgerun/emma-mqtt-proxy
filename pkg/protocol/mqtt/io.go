package mqtt

import (
	"io"
	"log"
)

// Inspired by bufio.Scanner, the mqtt.Scanner provides a convenient interface
// for reading a sequence of MQTT Packets from a stream.
type Scanner struct {
	r      io.Reader
	err    error
	packet *RawPacket
	done   bool
}

func NewScanner(reader io.Reader) *Scanner {
	return &Scanner{
		r: reader,
	}
}

func (s *Scanner) RawPacket() *RawPacket {
	return s.packet
}

// Err returns the first non-EOF error that was encountered by the Scanner.
func (s *Scanner) Err() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}

func (s *Scanner) setErr(err error) {
	if s.err == nil || s.err == io.EOF {
		s.err = err
	}
}

func (s *Scanner) Scan() bool {
	// TODO: this is heavily prototypical
	if s.done {
		return false
	}

	// read packet header
	log.Println("Reading next packet header")
	header := &PacketHeader{}
	_, err := ReadPacketHeader(s.r, header)
	if err != nil {
		log.Println("Error while reading packet header", err)
		s.setErr(err)
		s.done = true
		return false
	}
	log.Printf("Read packet type %s with remaining length %d\n", PacketTypeName(header.Type), header.Length)

	// allocate a buffer for the entire remaining packet
	rawPacket := NewRawPacket(*header)
	if header.Length == 0 {
		s.packet = rawPacket
		return true
	}

	// read the remaining packet fully
	log.Println("Reading remaining packet of size", rawPacket.Header.Length)
	_, err = io.ReadFull(s.r, rawPacket.Payload)
	if err != nil {
		log.Println("Error while reading packet header", err)
		s.setErr(err)
		s.done = true
		return false
	}

	s.packet = rawPacket
	log.Println("Successfully scanned packet", PacketTypeName(s.packet.Header.Type))
	return true
}
