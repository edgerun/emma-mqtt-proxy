package proxy

import (
	"github.com/edgerun/emma-mqtt-proxy/pkg/mqtt"
	"log"
	"net"
)

func readPackets(conn net.Conn, packets chan mqtt.Packet) {
	streamer := mqtt.NewStreamer(conn)

	log.Printf("reading packets from [%s]...\n", conn.RemoteAddr())
	for streamer.Next() {
		if streamer.Err() != nil {
			log.Printf("[%s] error while reading packet stream: %s\n", conn.RemoteAddr(), streamer.Err())
			break
		}

		packet := streamer.Packet()
		log.Printf("[%s] sent packet: %s\n", conn.RemoteAddr(), mqtt.PacketTypeName(packet.Type()))
		packets <- packet
	}

	if streamer.Err() != nil {
		log.Println("error while reading packet stream", streamer.Err())
	}

	log.Println("Exitting packet streamer")
}

func sendPackets(packets chan mqtt.Packet, conn net.Conn) (err error) {
	w := mqtt.NewWriter(conn)
	var n int64

	for packet := range packets {
		n, err = w.Write(packet)
		log.Printf("wrote %d bytes to [%s]\n", n, conn.RemoteAddr())

		if err != nil {
			return
		}
	}

	return
}
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
	errChan := make(chan error)

	defer close(clientPackets)
	defer close(brokerPackets)
	defer close(errChan)

	go func() {
		readPackets(clientConn, clientPackets)
	}()
	go func() {
		readPackets(brokerConn, brokerPackets)
	}()
	go func() {
		err := sendPackets(clientPackets, brokerConn)
		if err != nil {
			errChan <- err
			log.Println("Error sending packet to broker", err)
		}
	}()
	go func() {
		err := sendPackets(brokerPackets, clientConn)
		if err != nil {
			errChan <- err
			log.Println("Error sending packet to client", err)
		}
	}()

	<-errChan
	log.Println("breaking due to errors")
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
