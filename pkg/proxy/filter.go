package proxy

import "github.com/edgerun/emma-mqtt-proxy/pkg/mqtt"

type FilterAction int

const (
	ActionDropPacket FilterAction = iota
	ActionForwardPacket
	ActionProcessPacket
)

type HeaderFilter interface {
	onHeader(header *mqtt.PacketHeader) FilterAction
}

type PacketFilter interface {
	onPacket(packet mqtt.Packet)
}

type Filter interface {
	HeaderFilter
	PacketFilter
}
