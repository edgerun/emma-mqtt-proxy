package proxy

import (
	"github.com/edgerun/emma-mqtt-proxy/pkg/mqtt"
	"log"
	"net"
)

func startBridgeHandler(clientConn net.Conn) {
	brokerConn, err := net.Dial("tcp", "127.0.0.1:1884")
	if err != nil {
		log.Println("error dialing broker", brokerConn)
		clientConn.Close()
		return
	}

	bridge := NewBridge(clientConn, brokerConn)
	errors := bridge.Start()

	// example of how the bridge can be used to intercept packets and manipulate the routing
	bridge.SetRouterLeft(func(header *mqtt.PacketHeader) mqtt.Writer{
		log.Printf("client %s sent %s\n", clientConn.RemoteAddr(), header.Type)
		return bridge.SinkRight()
	})

	err = <-errors
	log.Println("first error:", err)

	brokerConn.Close()
	clientConn.Close()

	for err := range errors {
		log.Println("other errors:", err)
	}

	bridge.Wait()
}

func Serve(network string, address string) {
	ln, err := net.Listen(network, address)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("listening for connections on %s\n", ln.Addr())

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("accepted connection from %s\n", conn.RemoteAddr())
		go startBridgeHandler(conn)
	}
}
