package mqtt

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type DecodingStreamer struct {
	r io.Reader // the underlying stream

	// stateful packet processing internals
	limR   *io.LimitedReader // to limit the reading
	hBuf   *bytes.Buffer     // buffer used for the header
	buf    *bytes.Buffer     // buffer for the packet
	header *PacketHeader     // the last read packet header

	consumed bool // flag if packet has been consumed
}

func NewDecodingStreamer(r io.Reader) *DecodingStreamer {
	s := &DecodingStreamer{
		r:        r,
		limR:     &io.LimitedReader{R: r},
		hBuf:     bytes.NewBuffer(make([]byte, 5)),
		buf:      bytes.NewBuffer(make([]byte, 4096)),
		consumed: true,
	}

	return s
}

func (s *DecodingStreamer) ReadPacket() (Packet, error) {
	return s.DecodePacket()
}

func (s *DecodingStreamer) Next() (*PacketHeader, error) {
	if !s.consumed {
		return nil, StreamStateError
	}

	header, err := ReadHeaderFrom(s.r)
	if err != nil {
		return nil, err
	}

	s.header = header
	s.consumed = false

	return header, nil
}

// WriteTo circumvents packet decoding and instead writes the bytes of the current header and packet directly into the
// writer.
func (s *DecodingStreamer) WriteTo(w io.Writer) (n int64, err error) {
	if s.header == nil || s.consumed {
		return 0, StreamStateError
	}

	var nn int64

	nn, err = s.writeHeaderTo(w)
	n += nn

	nn, err = io.CopyN(w, s.r, int64(s.header.Length))
	n += nn

	if err != nil {
		return n, err
	}

	s.consumed = true
	return
}

func (s *DecodingStreamer) writeHeaderTo(w io.Writer) (int64, error) {
	s.hBuf.Reset()
	err := EncodeHeader(s.hBuf, s.header)
	if err != nil {
		return 0, err
	}
	return s.hBuf.WriteTo(w)
}

func (s *DecodingStreamer) DecodePacket() (Packet, error) {
	if s.header == nil || s.consumed {
		return nil, StreamStateError
	}

	buf := s.buf
	header := s.header
	r := s.limR

	r.N = int64(header.Length)
	// prepare buffer to read into
	buf.Reset()
	buf.Grow(int(header.Length)) // make sure we have enough space

	// read packet data into buffer
	n, err := buf.ReadFrom(r)
	if err != nil {
		return nil, err
	}
	if n != int64(header.Length) {
		panic("did not read complete packet")
	}

	// unmarshal packet into buffer (we've ensured that s.buf is exactly the remaining length of the MQTT packet)
	p, err := DecodePacket(buf, header)
	if err != nil {
		return nil, err
	}

	s.consumed = true
	return p, nil
}

func DecodePacket(buf *bytes.Buffer, h *PacketHeader) (p Packet, err error) {
	switch h.Type {
	case TypeConnect:
		p, err = DecodeConnectPacket(buf)
	case TypeConnAck:
		p, err = DecodeConnAckPacket(buf)
	case TypePublish:
		p, err = DecodePublishPacket(buf, h)
	case TypeSubscribe:
		p, err = DecodeSubscribePacket(buf)
	case TypeSubAck:
		p, err = DecodeSubAckPacket(buf)
	case TypePingReq:
		p, err = DecodePingReqPacket(buf)
	case TypePingResp:
		p, err = DecodePingRespPacket(buf)
	case TypeDisconnect:
		p, err = DecodeDisconnectPacket(buf)
	default:
		return nil, errors.New(fmt.Sprintf("unknown packet type %d", h.Type))
	}

	p.setHeader(h)

	return p, err
}

func ReadHeaderFrom(r io.Reader) (h *PacketHeader, err error) {
	// FIXME seems unnecessarily complicated
	buf := make([]byte, 5)

	n, err := r.Read(buf[:2])
	if err != nil {
		return
	}
	if n != 2 {
		// FIXME
		err = errors.New("error reading header")
		return
	}

	for i := 1; (buf[i]&cMask) > 0 && i < 5; i++ {
		n, err = r.Read(buf[i+1 : i+2])
		if err != nil {
			return
		}
	}

	h = &PacketHeader{}
	err = DecodeHeader(bytes.NewBuffer(buf), h)
	return
}

func DecodeHeader(buf *bytes.Buffer, h *PacketHeader) (err error) {
	typeAndFlags, err := buf.ReadByte()
	if err != nil {
		return
	}
	h.Type = PacketType(typeAndFlags >> 4)
	h.Flags = typeAndFlags & 0b00001111

	length, err := VariableByteUint32(buf)
	if err != nil {
		return
	}
	h.Length = length

	return
}

func DecodeConnAckPacket(buf *bytes.Buffer) (p *ConnAckPacket, err error) {
	p = &ConnAckPacket{}

	ackFlags, err := buf.ReadByte()
	if err != nil {
		return
	}
	p.SessionPresent = (ackFlags & 0x1) > 0

	p.ReturnCode, err = buf.ReadByte()
	if err != nil {
		return
	}

	return
}

func DecodeConnectPacket(buf *bytes.Buffer) (p *ConnectPacket, err error) {
	p = &ConnectPacket{}

	p.ProtocolName, err = LengthEncodedString(buf)

	switch p.ProtocolName {
	case "MQTT", "MQIsdp":
		break
	default:
		err = errors.New("unknown protocol type " + p.ProtocolName)
		return
	}

	// protocol level
	p.ProtocolLevel, err = buf.ReadByte()
	if err != nil {
		return
	}

	connectFlagByte, err := buf.ReadByte()
	if err != nil {
		return
	}
	p.ConnectFlags = DecodeConnectFlags(connectFlagByte)

	p.KeepAlive = Uint16(buf)

	if buf.Len() <= 0 {
		return
	}

	// TODO: read rest of the connect packet

	p.ClientId, err = LengthEncodedString(buf)
	if err != nil {
		return
	}

	return
}

func DecodeConnectFlags(b byte) ConnectFlags {
	return ConnectFlags{
		CleanSession: (b & 0x2) == 0x2,
		WillFlag:     (b & 0x4) == 0x4,
		WillQoS:      (b & 0x18) >> 3, // 0b00011000
		WillRetain:   (b & 0x20) == 0x20,
		PasswordFlag: (b & 0x40) == 0x40,
		UserNameFlag: (b & 0x80) == 0x80,
	}
}

func DecodePublishPacket(buf *bytes.Buffer, header *PacketHeader) (p *PublishPacket, err error) {
	p = &PublishPacket{}

	p.Dup = (header.Flags & 0b1000) > 0
	p.QoS = (header.Flags & 0b0110) >> 1
	p.Retain = (header.Flags & 0b0001) > 0

	varHeaderLen := 0
	p.TopicName, err = LengthEncodedString(buf)
	if err != nil {
		return
	}
	varHeaderLen += len(p.TopicName) + 2 // length encoded string field length
	if p.QoS > QoS0 {
		p.PacketId = Uint16(buf)
		varHeaderLen += 2
	}

	// copy the remaining length of the packet into the payload buffer
	remLen := int(header.Length) - varHeaderLen
	if remLen > 0 {
		p.Payload = make([]byte, remLen)
		_, err = buf.Read(p.Payload)
		if err != nil {
			return
		}
	}

	return
}

func decodeSubscription(buf *bytes.Buffer) (s Subscription, err error) {
	s = Subscription{}

	s.TopicFilter, err = LengthEncodedString(buf)
	if err != nil {
		return
	}
	qosByte, err := buf.ReadByte()
	if err != nil {
		return
	}
	s.QoS = qosByte & 0b00000011

	return
}

func DecodeSubscribePacket(buf *bytes.Buffer) (p *SubscribePacket, err error) {
	p = &SubscribePacket{}

	p.PacketId = Uint16(buf)

	var subs []Subscription

	// FIXME: read number of bytes specified in the header
	for buf.Len() > 0 {
		var sub Subscription
		sub, err = decodeSubscription(buf)
		if err != nil {
			return
		}
		subs = append(subs, sub)
	}

	p.Subscriptions = subs

	return
}

func DecodeSubAckPacket(buf *bytes.Buffer) (p *SubAckPacket, err error) {
	p = &SubAckPacket{}

	p.PacketId = Uint16(buf)

	// FIXME: read number of bytes specified in the header
	n := buf.Len()
	var codes = make([]SubAckCode, n)

	_, err = buf.Read(codes)
	if err != nil {
		return
	}

	p.ReturnCodes = codes

	return
}

func DecodePingRespPacket(_ *bytes.Buffer) (packet *PingRespPacket, err error) {
	packet = &PingRespPacket{}
	return
}

func DecodePingReqPacket(_ *bytes.Buffer) (packet *PingReqPacket, err error) {
	packet = &PingReqPacket{}
	return
}

func DecodeDisconnectPacket(_ *bytes.Buffer) (packet *DisconnectPacket, err error) {
	packet = &DisconnectPacket{}
	return
}
