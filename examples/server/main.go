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
		buf := make([]byte, 50)
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
		go binaryPrinter(conn)
	}
}
