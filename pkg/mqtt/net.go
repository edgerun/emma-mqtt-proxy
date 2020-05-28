package mqtt

import (
	"log"
	"net"
)

type PacketFilter = func(*RawPacket) bool

type ProxyConn struct {
	Client        net.Conn
	Broker        net.Conn
	ClientScanner *Scanner
	BrokerScanner *Scanner
}

func (proxy *ProxyConn) Close() {
	if proxy.Client != nil {
		proxy.Client.Close()
	}
	if proxy.Broker != nil {
		proxy.Broker.Close()
	}
}

func (proxy *ProxyConn) ForwardClientPackets(packetFilter PacketFilter) error {
	return ForwardPackets(proxy.Client, proxy.Broker, packetFilter)
}

func (proxy *ProxyConn) ForwardBrokerPackets(packetFilter PacketFilter) error {
	return ForwardPackets(proxy.Broker, proxy.Client, packetFilter)
}

func ForwardPackets(source net.Conn, sink net.Conn, packetFilter PacketFilter) error {
	scanner := NewScanner(source)

	for scanner.Scan() {
		packet := scanner.RawPacket()

		forward := packetFilter(packet)
		if forward {
			log.Printf("Writing packet %s to sink\n", PacketTypeName(packet.Header.Type))
			_, err := WriteRawPacket(sink, packet)
			if err != nil {
				log.Println("Error while writing packet", err)
				return err
			}
		}
	}

	return scanner.Err()
}

func DialBroker(client net.Conn, network string, address string) (*ProxyConn, error) {
	broker, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return NewProxyConn(client, broker), nil
}

func NewProxyConn(client net.Conn, broker net.Conn) *ProxyConn {
	return &ProxyConn{
		Client:        client,
		Broker:        broker,
		ClientScanner: NewScanner(client),
		BrokerScanner: NewScanner(broker),
	}
}
