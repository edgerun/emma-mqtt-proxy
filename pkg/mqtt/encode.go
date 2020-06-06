package mqtt

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type Encoder struct {
	w    io.Writer     // the underlying writer to write to
	hBuf *bytes.Buffer // buffer used for the header
	pBuf *bytes.Buffer // buffer used for the packet
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w:    w,
		hBuf: bytes.NewBuffer(make([]byte, 5)),
		pBuf: bytes.NewBuffer(make([]byte, 4096)),
	}
}

// Write the packet into the underlying io.Encoder. It does this as follows:
//  1. serialize the packet into a byte buffer to know how long it is
//  2. update the remaining length field of the header to the length that was written into byte buffer holding the packet
//  3. serialize the header into a byte buffer
//  4. write the header buffer
//  5. write the packet buffer
func (w *Encoder) Write(packet Packet) (err error) {
	// reset buffers
	hBuf := w.hBuf
	pBuf := w.pBuf
	hBuf.Reset()
	pBuf.Reset()

	// write packet into packet buffer
	err = EncodePacket(pBuf, packet)
	if err != nil {
		return
	}

	// create a new header for the outgoing packet that has as remaining length the length of the encoded packet (may be
	// different from the original incoming packet)
	h := &PacketHeader{
		Length: uint32(pBuf.Len()),
		Type:   packet.Type(),
		Flags:  packet.Flags(),
	}
	packet.setHeader(h) // does it really make sense to change the header of the original packet?

	// write header into header buffer
	err = EncodeHeader(hBuf, h)
	if err != nil {
		return
	}

	// write header buffer
	_, err = hBuf.WriteTo(w.w)
	if err != nil {
		return
	}

	// write packet buffer
	_, err = pBuf.WriteTo(w.w)
	if err != nil {
		return
	}

	return
}

// TODO: proper error handling

// Writes the packet into the buffer, but without the header.
func EncodePacket(buf *bytes.Buffer, p Packet) (err error) {
	switch p.Type() {
	case TypeConnect:
		return EncodeConnectPacket(buf, p.(*ConnectPacket))
	case TypeConnAck:
		return EncodeConnAckPacket(buf, p.(*ConnAckPacket))
	case TypePublish:
		return EncodePublishPacket(buf, p.(*PublishPacket))
	case TypeSubscribe:
		return EncodeSubscribePacket(buf, p.(*SubscribePacket))
	case TypeSubAck:
		return EncodeSubAckPacket(buf, p.(*SubAckPacket))
	case TypePingReq, TypePingResp, TypeDisconnect:
		return
	default:
		return errors.New(fmt.Sprintf("unknown packet type %d", p.Type()))
	}
}

func EncodeHeader(buf *bytes.Buffer, h *PacketHeader) (err error) {
	buf.WriteByte((byte(h.Type) << 4) | h.Flags)
	PutVariableByteUint32(buf, h.Length)
	return
}

func EncodeConnectPacket(buf *bytes.Buffer, p *ConnectPacket) (err error) {
	PutLengthEncodedString(buf, p.ProtocolName)
	buf.WriteByte(p.ProtocolLevel)
	_ = encodeConnectFlags(buf, p)
	PutUint16(buf, p.KeepAlive)
	PutLengthEncodedString(buf, p.ClientId)
	if p.WillFlag {
		PutLengthEncodedString(buf, p.WillTopic)
		PutLengthEncodedString(buf, p.WillMessage) // FIXME
	}
	if p.UserNameFlag {
		PutLengthEncodedString(buf, p.UserName)
	}
	if p.PasswordFlag {
		PutLengthEncodedField(buf, p.Password)
	}

	return
}

func encodeConnectFlags(buf *bytes.Buffer, p *ConnectPacket) (err error) {
	var flags byte

	if p.ConnectFlags.UserNameFlag {
		flags |= 0x80 // 0b10000000
	}
	if p.ConnectFlags.PasswordFlag {
		flags |= 0x40 // 0b01000000
	}
	if p.ConnectFlags.WillRetain {
		flags |= 0x20
	}

	flags |= p.ConnectFlags.WillQoS << 3

	if p.ConnectFlags.WillFlag {
		flags |= 0x4
	}
	if p.ConnectFlags.CleanSession {
		flags |= 0x2
	}

	buf.WriteByte(flags)
	return
}

func EncodeConnAckPacket(buf *bytes.Buffer, p *ConnAckPacket) (err error) {
	if p.SessionPresent {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}

	buf.WriteByte(p.ReturnCode)

	return
}

func EncodePublishPacket(buf *bytes.Buffer, p *PublishPacket) (err error) {
	PutLengthEncodedString(buf, p.TopicName)
	buf.Write(p.Payload)

	return
}

func EncodeSubscribePacket(buf *bytes.Buffer, p *SubscribePacket) (err error) {
	PutUint16(buf, p.PacketId)

	for _, sub := range p.Subscriptions {
		PutLengthEncodedString(buf, sub.TopicFilter)
		buf.WriteByte(sub.QoS)
	}

	return
}

func EncodeSubAckPacket(buf *bytes.Buffer, p *SubAckPacket) (err error) {
	PutUint16(buf, p.PacketId)

	for _, code := range p.ReturnCodes {
		buf.WriteByte(code & 0b10000011)
	}

	return
}
