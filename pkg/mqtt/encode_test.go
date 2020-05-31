package mqtt

import (
	"bytes"
	"testing"
)

func TestEncodeConnectPacket(t *testing.T) {
	expected := []byte{
		// connect packet
		// 16,   // connect + 0 flags  << HEADER
		// 29,   // remaining length   << HEADER
		0, 6, // protocol name length
		77, 81, 73, 115, 100, 112, // "MQIsdp"
		3,     // protocol level
		2,     // connect flags (X clean session)
		0, 60, // keepalive (60)
		0, 15, // client id length
		109, 111, 115, 113, 112, 117, 98, // mosqpub
		124, 57, 52, 48, 56, 45, 111, 109, // |9408-om
	}

	p := &ConnectPacket{
		ProtocolName:  "MQIsdp",
		ProtocolLevel: 3,
		ConnectFlags:  DecodeConnectFlags(2),
		KeepAlive:     60,
		ClientId:      "mosqpub|9408-om",
	}

	buf := bytes.NewBuffer(make([]byte, 4096))
	buf.Reset()

	err := EncodePacket(buf, p, TypeConnect)
	if err != nil {
		t.Error("unexpected error", err)
	}
	if buf.Len() != 29 {
		t.Error("unexpected encoded length", buf.Len())
	}

	actual := buf.Next(29)

	for i, b := range expected {
		if b != actual[i] {
			t.Errorf("unexpected value at index %d, %d != %d", i, actual[i], b)
		}
	}

}

func TestEncodePublishPacket(t *testing.T) {
	expected := []byte{
		// a publish packet
		//48, 10, // Header (publish)
		0, 4, // Topic length
		116, 101, 115, 116, // Topic (test)
		116, 101, 115, 116, // Payload (test),
	}

	p := &PublishPacket{
		TopicName:  "test",
		Payload: []byte("test"),
	}
	buf := bytes.NewBuffer(make([]byte, 4096))
	buf.Reset()

	err := EncodePacket(buf, p, TypePublish)
	if err != nil {
		t.Error("unexpected error", err)
	}
	if buf.Len() != 10 {
		t.Error("unexpected encoded length", buf.Len())
	}

	actual := buf.Next(len(expected))

	for i, b := range expected {
		if b != actual[i] {
			t.Errorf("unexpected value at index %d, %d != %d", i, actual[i], b)
		}
	}


}