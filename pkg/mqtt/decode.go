package mqtt

import "bytes"

func DecodeHeader(buf *bytes.Buffer) (h *PacketHeader, err error) {
	h = &PacketHeader{}

	typeAndFlags, err := buf.ReadByte()
	if err != nil {
		return
	}
	h.Type = typeAndFlags >> 4
	h.Flags = typeAndFlags & 0b00001111

	length, err := VariableByteUint32(buf)
	if err != nil {
		return
	}
	h.Length = length

	return
}

func DecodeConnectPacket(buf *bytes.Buffer) (p *ConnectPacket, err error) {
	p = &ConnectPacket{}

	p.ProtocolName, err = LengthEncodedString(buf)

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
		WillQoS:      b & 0x18, // 0b00011000
		WillRetain:   (b & 0x20) == 0x20,
		PasswordFlag: (b & 0x40) == 0x40,
		UserNameFlag: (b & 0x80) == 0x80,
	}
}
