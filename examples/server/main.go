package main

import (
	"github.com/edgerun/emma-mqtt-proxy/pkg/mqtt"
	"log"
	"net"
	"sync"
)

func readRawPackets(conn net.Conn, packets chan *mqtt.RawPacket) (err error) {
	defer conn.Close()

	scanner := mqtt.NewScanner(conn)

	for scanner.Scan() {
		packet := scanner.RawPacket()
		log.Printf("Received %s with length %d\n", mqtt.PacketTypeName(packet.Header.Type), packet.Header.Length)
		packets <- packet
	}

	return scanner.Err()
}

func writeRawPackets(packets chan *mqtt.RawPacket, conn net.Conn) (err error) {
	for packet := range packets {
		_, err = mqtt.WriteRawPacket(conn, packet)
		if err != nil {
			return
		}
	}

	return
}

func mqttProxyHandler(client net.Conn) {
	clientPacketChan := make(chan *mqtt.RawPacket)
	brokerPacketChan := make(chan *mqtt.RawPacket)

	defer client.Close()
	var wg sync.WaitGroup

	// initialize the client packet read loop
	wg.Add(1)
	go func() {
		err := readRawPackets(client, clientPacketChan)
		if err != nil {
			log.Println("error: while reading packets", err)
		}
		wg.Done()
	}()

	connectPacket := <-clientPacketChan
	if connectPacket.Header.Type != mqtt.TypeConnect {
		log.Println("error: eirst packet was not a connect packet")
		return
	}

	broker, err := net.Dial("tcp", "127.0.0.1:1883")
	if err != nil {
		log.Println("error: could not establish connection to broker", err)
		return
	}
	// forward the connect packet, and then establish a pipe
	_, err = mqtt.WriteRawPacket(broker, connectPacket)
	if err != nil {
		log.Println("error: could not establish connection to broker", err)
		return
	}

	// forward all packets from the client to the broker
	wg.Add(1)
	go func() {
		err := writeRawPackets(clientPacketChan, broker)
		if err != nil {
			log.Println("error: while writing packets to broker", err)
		}
		wg.Done()
	}()

	// start reading on broker
	wg.Add(1)
	go func() {
		err := readRawPackets(broker, brokerPacketChan)
		if err != nil {
			log.Println("error: while writing packets to broker", err)
		}
		wg.Done()
	}()

	// forward all packets from the broker to the client
	wg.Add(1)
	go func() {
		err := writeRawPackets(brokerPacketChan, client)
		if err != nil {
			log.Println("error: while writing packets to client", err)
		}
		wg.Done()
	}()

	log.Println("waiting for proxy connection to end")
	wg.Wait()
	log.Println("exitting MqttProxyHandler")
}

func binaryPrinter(conn net.Conn) {
	for {
		buf := make([]byte, 4096)
		read, err := conn.Read(buf)
		if err != nil {
			log.Println("Error while reading from stream", err)
			break
		}
		log.Printf("Read %d bytes\n", read)
		rbuf := buf[:(read + 1)]
		for i, b := range rbuf {
			log.Printf("%d: %d\n", i, b)
		}
	}
}

func readPackets(conn net.Conn, packets chan mqtt.Packet) {
	streamer := mqtt.NewStreamer(conn)

	log.Printf("reading packets from [%s]...\n", conn.RemoteAddr())
	for streamer.Next() {
		if streamer.Err() != nil {
			log.Printf("[%s] error while reading packet stream: %s\n", conn.RemoteAddr(), streamer.Err())
			break
		}

		packet := streamer.Packet()
		log.Printf("[%s] sent packet: %s\n", conn.RemoteAddr(), mqtt.PacketTypeName(packet.Header().Type))
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

func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:1883")
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
