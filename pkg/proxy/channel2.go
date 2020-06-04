package proxy

import (
	"bytes"
	"fmt"
	"github.com/edgerun/emma-mqtt-proxy/pkg/mqtt"
	"io"
)

const defaultBufSize = 4096

const cMask = 0b10000000 // 128 -- mask for continuation bit in variable integer
const dMask = 0b01111111 // 127 -- mask for data in variable integer

type Streamer struct {
	r      io.Reader          // the underlying stream
	limR   *io.LimitedReader  // limited reader for reading packet payloads into buffer from stream
	buf    *bytes.Buffer      // the buffer that holds packet data to be unmarshalled
	header *mqtt.PacketHeader // current packet header
	packet mqtt.Packet        // current packet
	err    error              // the first non-eof error that occurred
	done   bool               // if we're done
	filter HeaderFilter
}

func NewStreamer(r io.Reader) *Streamer {
	buf := bytes.NewBuffer(make([]byte, defaultBufSize))
	buf.Reset()

	return &Streamer{
		r:    r,
		limR: &io.LimitedReader{R: r, N: 0},
		buf:  buf,
	}
}

func (s *Streamer) RegisterFilter(filter HeaderFilter) {
	s.filter = filter
}

func (s *Streamer) Packet() mqtt.Packet {
	return s.packet
}

func (s *Streamer) Err() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}

func (s *Streamer) setErr(err error) {
	if s.err == nil || s.err == io.EOF {
		s.err = err
	}
}

func (s *Streamer) bail(err error) bool {
	s.setErr(err)
	s.done = true
	return false
}

func (s *Streamer) Next() bool {
	if s.done {
		return false
	}

	for {
		action := ActionForwardPacket
		r := s.r

		// read the packet header
		header, err := mqtt.ReadHeaderFrom(r)
		if err != nil {
			return s.bail(err)
		}

		if s.filter != nil {
			action = s.filter.onHeader(header)
		}
		// TODO: apply filters and potentially drop packet, or skip decoding and copy data directly

		switch action {
		case ActionDropPacket:
			err = s.dropPacket(header)
			if err != nil {
				return s.bail(err)
			}
			continue
		case ActionForwardPacket:
			// TODO: io.Copy
			err = s.forwardPacket(header)
			if err != nil {
				return s.bail(err)
			}
			continue
		case ActionProcessPacket:
			// read the packet
			s.packet, err = s.processPacket(header)
			if err != nil {
				return s.bail(err)
			}
			return true
		default:
			panic("unknown packet action")
		}
	}
}

func (s *Streamer) dropPacket(header *mqtt.PacketHeader) (err error) {
	// TODO
	return
}

func (s *Streamer) forwardPacket(header *mqtt.PacketHeader) (err error) {
	// TODO
	s.limR.N = int64(header.Length)
	return
}

func (s *Streamer) processPacket(header *mqtt.PacketHeader) (p mqtt.Packet, err error) {
	// prepare limited reader to read the packet length exactly from the underlying reader
	if s.buf.Len() != 0 {
		panic(fmt.Sprintf("packet buffer should be empty, was %d", s.buf.Len()))
	}
	s.limR.N = int64(header.Length)

	// prepare buffer to read into
	s.buf.Reset()
	s.buf.Grow(int(header.Length)) // make sure we have enough space

	// read packet data into buffer
	_, err = s.buf.ReadFrom(s.limR)
	if err != nil {
		return
	}

	// unmarshal packet into buffer (we've ensured that s.buf is exactly the remaining length of the MQTT packet)
	p, err = mqtt.DecodePacket(s.buf, header)
	if err != nil {
		return
	}
	return
}
