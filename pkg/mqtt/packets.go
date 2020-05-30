package mqtt

type PacketType = uint8 // uint4
type QoS = byte

const MaxPacketSize = 268435455 // packet size is stored in a variable byte integer with max 4 bytes: (2^(7*4)) - 1

const (
	QoS1 QoS = 0
	QoS2 QoS = 1
	QoS3 QoS = 2
)

const (
	TypeReserved    PacketType = 0  // not a real packet type, type is reserved
	TypeConnect     PacketType = 1  // Client to Server	Connection request
	TypeConnAck     PacketType = 2  // Server to Client	Connect acknowledgment
	TypePublish     PacketType = 3  // Client to Server or Server to Client	Publish message
	TypePubAck      PacketType = 4  // Client to Server or Server to Client	Publish acknowledgment (QoS 1)
	TypePubRec      PacketType = 5  // Client to Server or Server to Client	Publish received (QoS 2 delivery part 1)
	TypePubRel      PacketType = 6  // Client to Server or Server to Client	Publish release (QoS 2 delivery part 2)
	TypePubComp     PacketType = 7  // Client to Server or Server to Client	Publish complete (QoS 2 delivery part 3)
	TypeSubscribe   PacketType = 8  // Client to Server	Subscribe request
	TypeSubAck      PacketType = 9  // Server to Client	Subscribe acknowledgment
	TypeUnsubscribe PacketType = 10 // Client to Server	Unsubscribe request
	TypeUnsubAck    PacketType = 11 // Server to Client	Unsubscribe acknowledgment
	TypePingReq     PacketType = 12 // Client to Server	PING request
	TypePingResp    PacketType = 13 // Server to Client	PING response
	TypeDisconnect  PacketType = 14 // Client to Server or Server to Client	Disconnect notification
	TypeAuth        PacketType = 15 // Client to Server or Server to Client	Authentication exchange
)

var packetTypeNames = map[PacketType]string{
	TypeReserved:    "Reserved",
	TypeConnect:     "CONNECT",
	TypeConnAck:     "CONNACK",
	TypePublish:     "PUBLISH",
	TypePubAck:      "PUBACK",
	TypePubRec:      "PUBREC",
	TypePubRel:      "PUBREL",
	TypePubComp:     "PUBCOMP",
	TypeSubscribe:   "SUBSCRIBE",
	TypeSubAck:      "SUBACK",
	TypeUnsubscribe: "UNSUBSCRIBE",
	TypeUnsubAck:    "UNSUBACK",
	TypePingReq:     "PINGREQ",
	TypePingResp:    "PINGRESP",
	TypeDisconnect:  "DISCONNECT",
	TypeAuth:        "AUTH",
}

func PacketTypeName(packetType PacketType) string {
	v, ok := packetTypeNames[packetType]
	if !ok {
		return "UNKNOWN"
	}
	return v
}

type Flags = uint8 // uint4

// +-----------------------------------+
// |   | 7 | 6 | 5 | 4 | 3 | 2 | 1 | 0 |
// +-----------------------------------+
// | 0 |  PACKET TYPE  |     FLAGS     |
// +-----------------------------------+
// | 1 | 1     VARIABLE LENGTH INT     | MSB is the 'continuation bit', and the 7 LSB contain the next bits of the int.
// | 2 | 1            ...              | 0 in the MSG indicates that the variable length int is done. The maximum number
// |...| 0            ...              | of bytes in the Variable Byte Integer field is four: so 4*7 bit = max 28 bit
// +-----------------------------------+
type PacketHeader struct {
	Type   PacketType
	Flags  Flags
	Length uint32
}

type Packet interface {
	Header() *PacketHeader
}

type packet struct {
	header *PacketHeader
}

type RawPacket struct {
	Header  PacketHeader
	Payload []byte
}

type ConnectFlags struct {
	CleanSession bool
	WillFlag     bool
	WillQoS      QoS
	WillRetain   bool
	PasswordFlag bool
	UserNameFlag bool
}

type ConnectPacket struct {
	packet
	ConnectFlags
	ProtocolName  string
	ProtocolLevel uint8
	KeepAlive     uint16
	ClientId      string
	WillTopic     string
	WillMessage   string // FIXME: should be a []byte field
	UserName      string
	Password      []byte
}

type ConnAckPacket struct {
	packet
	Flags      byte
	ReturnCode byte
}

func (p *packet) Header() *PacketHeader {
	return p.header
}

func NewRawPacket(header PacketHeader) *RawPacket {
	return &RawPacket{
		Header: header,
	}
}

func (h *PacketHeader) packetByteSize() int {
	return int(h.Length) + h.byteSize()
}

func (h *PacketHeader) byteSize() int {
	varInt := h.Length

	if varInt <= 127 {
		return 2
	}
	if varInt <= 16383 {
		return 3
	}
	if varInt <= 2097151 {
		return 4
	}
	if varInt <= 268435455 {
		return 5
	}

	panic("Header too large")
}
