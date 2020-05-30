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
	}

	println(input) // TODO
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
