package mqtt

import (
	"bytes"
	"testing"
)

func TestStreamer_ParseConnectPacket(t *testing.T) {
	reader := bytes.NewReader([]byte{
		// connect packet
		16, 31,
		0, 6, 77, 81, 73, 115, 100, 112,
		3, 2, 0, 60, 0, 17, 109, 111, 115,
		113, 112, 117, 98, 124, 54, 56, 53, 52,
		45, 116, 121, 116, 111,
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
	assertStringEquals(t, "mosqpub|6854-tyto", cp.ClientId)
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