package mqtt

type FilterAction int

const (
	ActionDropPacket FilterAction = iota
	ActionForwardPacket
	ActionProcessPacket
)

type HeaderFilter interface {
	onHeader(header *PacketHeader) FilterAction
}

type PacketFilter interface {
	onPacket(packet Packet)
}

type Filter interface {
	HeaderFilter
	PacketFilter
}
