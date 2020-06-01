package mqtt

import (
	"bytes"
	"errors"
	"fmt"
)

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
	case TypePingReq, TypePingResp:
		return
	default:
		return errors.New(fmt.Sprintf("unknown packet type %d", p.Type()))
	}
}

func EncodeHeader(buf *bytes.Buffer, h *PacketHeader) (err error) {
	buf.WriteByte((h.Type << 4) | h.Flags)
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
