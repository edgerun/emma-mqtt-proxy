package filter

import (
	"fmt"
	"github.com/edgerun/emma-mqtt-proxy/pkg/mqtt"
)

type Action int

const (
	Next    Action = 0
	Forward Action = 1
	Drop    Action = 2
)

type Chain struct {
	headerFilter []HeaderFilter
	packetFilter []PacketFilter
}

type HeaderFilter interface {
	FilterHeader(header *mqtt.PacketHeader) Action
}

type PacketFilter interface {
	FilterPacket(packet *mqtt.RawPacket) Action
}

type Printer struct {
}

func (Printer) FilterHeader(header *mqtt.PacketHeader) Action {
	fmt.Printf("packet header: %s[%d]\n", mqtt.PacketTypeName(header.Type), header.Length)
	return Next
}

func (Printer) FilterPacket(packet *mqtt.RawPacket) Action {
	fmt.Printf("packet: %s[%d]\n", mqtt.PacketTypeName(packet.Header.Type), len(packet.Payload))
	return Next
}
