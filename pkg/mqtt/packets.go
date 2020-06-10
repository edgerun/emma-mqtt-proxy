package mqtt

type PacketType uint8 // uint4
type QoS = byte
type SubAckCode = byte

const MaxPacketSize = 268435455 // packet size is stored in a variable byte integer with max 4 bytes: (2^(7*4)) - 1

const (
	QoS0 QoS = 0x00
	QoS1 QoS = 0x01
	QoS2 QoS = 0x02
)

const (
	MaxQoS0 SubAckCode = 0x00
	MaxQoS1 SubAckCode = 0x01
	MaxQoS2 SubAckCode = 0x02
	Failure SubAckCode = 0x80
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
		panic("invalid packet type")
	}
	return v
}

func (t PacketType) String() string {
	return PacketTypeName(t)
}

type Flags = uint8 // uint4

// The fixed header of an MQTT protocol holds the packet type, fixed packet-specific flags, and the remaining length
// in bytes of the packet encoded in a max 4 byte encoded variable integer.
// The structure is as follows:
//     +-----------------------------------+
//     |   | 7 | 6 | 5 | 4 | 3 | 2 | 1 | 0 |
//     +-----------------------------------+
//     | 0 |  PACKET TYPE  |     FLAGS     |
//     +-----------------------------------+
//     | 1 | 1     VARIABLE LENGTH INT     | MSB is the 'continuation bit', and the 7 LSB contain the next bits of the int.
//     | 2 | 1            ...              | 0 in the MSG indicates that the variable length int is done. The maximum number
//     |...| 0            ...              | of bytes in the Variable Byte Integer field is four: so 4*7 bit = max 28 bit
//     +-----------------------------------+
//
// After a packet has been deserialized, the value of Length becomes meaningless, as it can change once
type PacketHeader struct {
	Type   PacketType
	Flags  Flags
	Length uint32
}

type Packet interface {
	// Type returns the constant identifying this particular packet's type.
	Type() PacketType
	// Returns the static header flags. This is 0 in most cases.
	Flags() Flags

	Header() *PacketHeader
	setHeader(header *PacketHeader)
}

type headerContainer struct {
	header *PacketHeader
}

func (p *headerContainer) Header() *PacketHeader {
	return p.header
}

func (p *headerContainer) setHeader(header *PacketHeader) {
	p.header = header
}

type ConnectPacket struct {
	headerContainer
	ConnectFlags
	ProtocolName  string
	ProtocolLevel uint8
	KeepAlive     uint16
	ClientId      string
	WillTopic     string
	WillMessage   []byte
	UserName      string
	Password      []byte
}

type ConnectFlags struct {
	CleanSession bool
	WillFlag     bool
	WillQoS      QoS
	WillRetain   bool
	PasswordFlag bool
	UserNameFlag bool
}

func (*ConnectPacket) Type() PacketType {
	return TypeConnect
}

func (*ConnectPacket) Flags() Flags {
	return 0
}

type ConnAckPacket struct {
	headerContainer
	SessionPresent bool
	ReturnCode     byte
}

func (*ConnAckPacket) Type() PacketType {
	return TypeConnAck
}

func (*ConnAckPacket) Flags() Flags {
	return 0
}

type PingReqPacket struct {
	headerContainer
}

func (*PingReqPacket) Type() PacketType {
	return TypePingReq
}

func (*PingReqPacket) Flags() Flags {
	return 0
}

type PingRespPacket struct {
	headerContainer
}

func (*PingRespPacket) Type() PacketType {
	return TypePingResp
}

func (*PingRespPacket) Flags() Flags {
	return 0
}

type PublishPacket struct {
	headerContainer
	// unmarshalled static header flags
	Dup    bool
	QoS    QoS
	Retain bool
	// variable header + payload
	TopicName string
	PacketId  uint16
	Payload   []byte
}

func (*PublishPacket) Type() PacketType {
	return TypePublish
}

// Fixed header flags for the publish packet:
//
//     +--------+--------+--------+--------+
//     | 0      | 1      | 2      | 3      |
//     +--------+-----------------+--------+
//     | DUP    |       QoS       | RETAIN |
//     +--------+-----------------+--------+
func (p *PublishPacket) Flags() (flags Flags) {

	flags = 0
	if p.Retain {
		flags |= 0x1
	}
	flags |= p.QoS << 1
	if p.Dup {
		flags |= 0x8
	}
	return
}

type PubAckPacket struct {
	headerContainer
	PacketId uint16
}

func (p PubAckPacket) Type() PacketType {
	return TypePubAck
}

func (p PubAckPacket) Flags() Flags {
	return 0
}

type PubRecPacket struct {
	headerContainer
	PacketId uint16
}

func (p PubRecPacket) Type() PacketType {
	return TypePubRec
}

func (p PubRecPacket) Flags() Flags {
	return 0
}

type PubRelPacket struct {
	headerContainer
	PacketId uint16
}

func (p PubRelPacket) Type() PacketType {
	return TypePubRel
}

// Fixed header flags for the PUBREL packet:
//
//       +--------+--------+--------+--------+
// bit   | 3      | 2      | 1      | 0      |
//       +--------+-----------------+--------+
// value | 0      | 0      | 1      | 0      |
//       +--------+-----------------+--------+
func (p PubRelPacket) Flags() Flags {
	return 2
}

type PubCompPacket struct {
	headerContainer
	PacketId uint16
}

func (p PubCompPacket) Type() PacketType {
	return TypePubComp
}

func (p PubCompPacket) Flags() Flags {
	return 0
}

type SubscribePacket struct {
	headerContainer
	PacketId      uint16
	Subscriptions []Subscription
}

type Subscription struct {
	TopicFilter string
	QoS         QoS
}

func (*SubscribePacket) Type() PacketType {
	return TypeSubscribe
}

func (*SubscribePacket) Flags() Flags {
	return 0
}

type SubAckPacket struct {
	headerContainer
	PacketId    uint16
	ReturnCodes []SubAckCode
}

func (*SubAckPacket) Type() PacketType {
	return TypeSubAck
}

func (*SubAckPacket) Flags() Flags {
	return 0
}

type UnsubscribePacket struct {
	headerContainer
	PacketId     uint16
	TopicFilters []string
}

func (p UnsubscribePacket) Type() PacketType {
	return TypeUnsubscribe
}

// Fixed header flags for the SUBSCRIBE packet:
//
//       +--------+--------+--------+--------+
// bit   | 3      | 2      | 1      | 0      |
//       +--------+-----------------+--------+
// value | 0      | 0      | 1      | 0      |
//       +--------+-----------------+--------+
func (p UnsubscribePacket) Flags() Flags {
	return 2
}

type UnsubAckPacket struct {
	headerContainer
	PacketId uint16
}

func (p UnsubAckPacket) Type() PacketType {
	return TypeUnsubAck
}

func (p UnsubAckPacket) Flags() Flags {
	return 0
}

type DisconnectPacket struct {
	headerContainer
}

func (*DisconnectPacket) Type() PacketType {
	return TypeDisconnect
}

func (*DisconnectPacket) Flags() Flags {
	return 0
}
