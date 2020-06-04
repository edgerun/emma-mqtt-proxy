package mqtt

import (
	"bytes"
	"fmt"
	"io"
)

const defaultBufSize = 4096

type Scanner struct {
	r      io.Reader         // the underlying stream
	limR   *io.LimitedReader // limited reader for reading packet payloads into buffer from stream
	buf    *bytes.Buffer     // the buffer that holds packet data to be unmarshalled
	header *PacketHeader     // current packet header
	packet Packet            // current packet
	err    error             // the first non-eof error that occurred
	done   bool              // if we're done
}

func NewScanner(r io.Reader) *Scanner {
	buf := bytes.NewBuffer(make([]byte, defaultBufSize))
	buf.Reset()

	return &Scanner{
		r:    r,
		limR: &io.LimitedReader{R: r, N: 0},
		buf:  buf,
	}
}

func (s *Scanner) Packet() Packet {
	return s.packet
}

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

func (s *Scanner) bail(err error) bool {
	s.setErr(err)
	s.done = true
	return false
}

func (s *Scanner) Next() bool {
	if s.done {
		return false
	}

	r := s.r

	// read the packet header
	header, err := ReadHeaderFrom(r)
	if err != nil {
		return s.bail(err)
	}

	// TODO: apply filters and potentially drop packet, or skip decoding and copy data directly

	// read the packet
	s.packet, err = s.readPacket(header)
	if err != nil {
		return s.bail(err)
	}

	return true
}

func (s *Scanner) readPacket(header *PacketHeader) (p Packet, err error) {
	// prepare limited reader to read the packet length exactly from the underlying reader
	if s.buf.Len() != 0 {
		panic(fmt.Sprintf("packet buffer should be empty, was %d", s.buf.Len()))
	}
	s.limR.N = int64(header.Length)

	// prepare buffer to read into
	s.buf.Reset()
	s.buf.Grow(int(header.Length))  // make sure we have enough space

	// read packet data into buffer
	_, err = s.buf.ReadFrom(s.limR)
	if err != nil {
		return
	}

	// unmarshal packet into buffer (we've ensured that s.buf is exactly the remaining length of the MQTT packet)
	p, err = DecodePacket(s.buf, header)
	if err != nil {
		return
	}
	return
}
