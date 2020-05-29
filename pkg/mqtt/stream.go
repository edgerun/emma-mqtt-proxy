package mqtt

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

const defaultBufSize = 4096

type Streamer struct {
	r      io.Reader         // the underlying stream
	limR   *io.LimitedReader // limited reader for reading packet payloads into buffer from stream
	buf    *bytes.Buffer     // the buffer that holds packet data to be unmarshalled
	header *PacketHeader     // current packet header
	packet Packet            // current packet
	err    error             // the first non-eof error that occurred
	done   bool              // if we're done
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

func (s *Streamer) Packet() Packet {
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

	r := s.r

	// read the packet header
	var header = &PacketHeader{}
	_, err := ReadPacketHeader(r, header)
	if err != nil {
		return s.bail(err)
	}

	// TODO: apply filters and potentially drop or pipe packet

	// prepare limited reader
	if s.buf.Len() != 0 {
		panic(fmt.Sprintf("packet buffer should be empty, was %d", s.buf.Len()))
	}
	s.buf.Reset()
	s.limR.N = int64(header.Length)
	s.buf.Grow(int(header.Length))  // make sure we have enough space
	_, err = s.buf.ReadFrom(s.limR) // read packet data into buffer TODO: what happens on partial reads?
	if err != nil {
		return s.bail(err)
	}

	// unmarshal packet
	packet, err := s.readPacket(header)
	s.packet = packet
	if err != nil {
		return s.bail(err)
	}

	return true
}

func (s *Streamer) readPacket(header *PacketHeader) (Packet, error) {
	switch header.Type {
	case TypeConnect:
		return s.readConnect(header)
	default:
		return nil, errors.New("unknown packet type " + PacketTypeName(header.Type))
	}

}

func NewConnectFlags(b byte) ConnectFlags {
	// TODO: deserialize flags
	return ConnectFlags{
		CleanSession: false,
		WillFlag:     false,
		WillQoS:      0,
		WillRetain:   false,
		PasswordFlag: false,
		UserNameFlag: false,
	}
}

func (s *Streamer) readConnect(header *PacketHeader) (packet *ConnectPacket, err error) {
	packet = &ConnectPacket{}
	packet.header = header

	buf := s.buf // buffer contains the entire packet

	// protocol name
	packet.ProtocolName, err = LengthEncodedString(buf)
	if err != nil {
		return
	}
	switch packet.ProtocolName {
	case "MQTT":
	case "MQIsdp":
		break
	default:
		err = errors.New("unknown protocol type " + packet.ProtocolName)
		return
	}

	// protocol level
	packet.ProtocolLevel, err = buf.ReadByte()
	if err != nil {
		return
	}

	connectFlagByte, err := buf.ReadByte()
	if err != nil {
		return
	}
	packet.ConnectFlags = NewConnectFlags(connectFlagByte)

	packet.KeepAlive = Uint16(buf)

	if buf.Len() <= 0 {
		return
	}

	// TODO: read rest of the connect packet
	packet.ClientId, err = LengthEncodedString(buf)
	if err != nil {
		return
	}

	return
}
