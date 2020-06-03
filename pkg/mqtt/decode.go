package mqtt

import (
	"bytes"
	"errors"
	"fmt"
)

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
	default:
		return nil, errors.New(fmt.Sprintf("unknown packet type %d", h.Type))
	}

	p.setHeader(h)

	return p, err
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

func DecodePingRespPacket(buf *bytes.Buffer) (packet *PingRespPacket, err error) {
	packet = &PingRespPacket{}
	return
}

func DecodePingReqPacket(buf *bytes.Buffer) (packet *PingReqPacket, err error) {
	packet = &PingReqPacket{}
	return
}
