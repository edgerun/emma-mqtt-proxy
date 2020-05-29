package mqtt

import "net"

type ClientSession struct {
	LocalAddr  net.Addr
	RemoteAddr net.Addr
	ConnInfo   ConnectPacket
}
