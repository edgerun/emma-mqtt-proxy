package proxy

import (
	"github.com/edgerun/emma-mqtt-proxy/pkg/mqtt"
	"log"
	"net"
)

func startProxyHandler(clientConn net.Conn) {
	defer clientConn.Close()

	brokerConn, err := net.Dial("tcp", "127.0.0.1:1884")
	if err != nil {
		log.Println("error dialing broker", brokerConn)
		return
	}
	defer brokerConn.Close()

	clientPackets := make(chan mqtt.Packet)
	brokerPackets := make(chan mqtt.Packet)

	clientChan := NewChannel()
	clientChan.SetConnection(clientConn)
	clientChan.SetPacketListener(clientPackets)

	brokerChan := NewChannel()
	brokerChan.SetConnection(brokerConn)
	brokerChan.SetPacketListener(brokerPackets)

	go func() {
		for p := range clientPackets {
			log.Println("sending packet to broker", p.Type())
			brokerChan.Send(p)
		}
	}()
	go func() {
		for p := range brokerPackets {
			log.Println("sending packet to client", p.Type())
			clientChan.Send(p)
		}
	}()

	go brokerChan.Run()

	err = clientChan.Run()

	close(brokerPackets)
	close(clientPackets)

	brokerChan.Stop()
	clientChan.Stop()
}

func Serve(network string, address string) {
	ln, err := net.Listen(network, address)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Accept connection on port")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Calling connection handler")
		go startProxyHandler(conn)
	}
}
