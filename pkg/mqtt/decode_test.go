package mqtt

import (
	"bytes"
	"testing"
)

func TestDecodePublishPacket(t *testing.T) {
	input := []byte{
		// a publish packet
		48, 10, // Header (publish)
		0, 4, // Topic length
		116, 101, 115, 116, // Topic (test)
		116, 101, 115, 116, // Payload (test),
		1, 1, 1, // superfluous bytes to verify packet is processed correctly
	}

	buf := bytes.NewBuffer(input)

	header, err := DecodeHeader(buf)
	if err != nil {
		t.Error("unexpected error", err)
	}
	packet, err := DecodePublishPacket(buf, header)
	if err != nil {
		t.Error("unexpected error", err)
	}

	if packet.TopicName != "test" {
		t.Error("unexpected topic name", packet.TopicName)
	}
	if string(packet.Payload) != "test" {
		t.Error("unexpected payload string", string(packet.Payload))
	}
}

func TestDecodeConnectPacket(t *testing.T) {
	reader := bytes.NewReader([]byte{
		// connect packet
		16,   // connect + 0 flags
		29,   // remaining length
		0, 6, // protocol name length
		77, 81, 73, 115, 100, 112, // "MQIsdp"
		3,     // protocol level
		2,     // connect flags (X clean session)
		0, 60, // keepalive (60)
		0, 15, // client id length
		109, 111, 115, 113, 112, 117, 98, // mosqpub
		124, 57, 52, 48, 56, 45, 111, 109, // |9408-om
	})

	buf := bytes.NewBuffer(make([]byte, 4096))
	buf.Reset()

	_, err := buf.ReadFrom(reader)
	if err != nil {
		t.Error("unexpected error", err)
	}

	h, err := DecodeHeader(buf)
	if err != nil {
		t.Error("unexpected error", err)
	}
	if h == nil {
		t.Error("decode header returned nil")
		return
	}
	if h.Length != 29 {
		t.Error("Unexpected remaining length", h.Length)
	}

	p, err := DecodeConnectPacket(buf)
	if err != nil {
		t.Error("unexpected error", err)
	}
	if p == nil {
		t.Error("decode packet returned nil")
		return
	}

	if p.ProtocolName != "MQIsdp" {
		t.Error("unexpected protocol name", p.ProtocolName)
	}
	if p.KeepAlive != 60 {
		t.Error("unexpected keepalive value", p.KeepAlive)
	}
	if p.ProtocolLevel != 3 {
		t.Error("unexpected protocol level", p.ProtocolLevel)
	}
	if p.ClientId != "mosqpub|9408-om" {
		t.Error("unexpected client id", p.ClientId)
	}

}

func TestDecodeSubscribePacket(t *testing.T) {
	reader := bytes.NewReader([]byte{
		// subscribe packet
		130, // subscribe + 0010
		14,  // remaining length
		// variable header
		0, 42, // packet ID
		// ## payload
		// ### topic filter "a/b"
		0, 3, // length
		0x61, 0x2F, 0x62, // a/b
		// ### topic filter "c/d"
		1,    // request QoS (1)
		0, 3, // length
		0x63, 0x2F, 0x64, // c/d
		2, // request QoS (2)
	})

	buf := bytes.NewBuffer(make([]byte, 4096))
	buf.Reset()
	buf.ReadFrom(reader)

	header, err := DecodeHeader(buf)
	if err != nil {
		t.Error("unexpected error", err)
	}
	if header.Type != TypeSubscribe {
		t.Error("unexpected packet type", header.Type)
	}

	packet, err := DecodeSubscribePacket(buf)
	if err != nil {
		t.Error("unexpected error", err)
	}

	if packet.PacketId != 42 {
		t.Error("unexpected packet id", packet.PacketId)
	}
	if len(packet.Subscriptions) != 2 {
		t.Error("unexpected number of subscriptions", len(packet.Subscriptions))
	}

	if packet.Subscriptions[0].TopicFilter != "a/b" {
		t.Error("unexpected topic filter", packet.Subscriptions[0].TopicFilter)
	}
	if packet.Subscriptions[0].QoS != QoS1 {
		t.Error("unexpected QoS level", packet.Subscriptions[0].QoS)
	}

	if packet.Subscriptions[1].TopicFilter != "c/d" {
		t.Error("unexpected topic filter", packet.Subscriptions[1].TopicFilter)
	}
	if packet.Subscriptions[1].QoS != QoS2 {
		t.Error("unexpected QoS level", packet.Subscriptions[1].QoS)
	}

}
