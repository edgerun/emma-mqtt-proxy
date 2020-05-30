package mqtt

import (
	"bytes"
	"testing"
)

func TestStreamer_ParseConnectPacket(t *testing.T) {
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

	streamer := NewStreamer(reader)

	streamer.Next()

	if streamer.Err() != nil {
		t.Error("unexpected error", streamer.Err())
	}
	p := streamer.Packet()

	if p == nil {
		t.Error("expected packet to be not nil")
		return
	}

	header := p.Header()
	if header.Type != TypeConnect {
		t.Errorf("expected packet type %s, got %s", "CONNECT", PacketTypeName(header.Type))
	}
	cp := p.(*ConnectPacket)

	assertIntEquals(t, 60, int(cp.KeepAlive))
	assertStringEquals(t, "mosqpub|9408-om", cp.ClientId)
}

func assertIntEquals(t *testing.T, expected int, actual int) {
	if expected != actual {
		t.Errorf("%d != %d", actual, expected)
	}
}
func assertStringEquals(t *testing.T, expected string, actual string) {
	if expected != actual {
		t.Errorf("%s != %s", actual, expected)
	}
}
