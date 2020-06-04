package mqtt

import (
	"bytes"
	"testing"
)

func TestWriterStreamerIntegration(t *testing.T) {
	expected := []byte{
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
	}

	streamer := NewScanner(bytes.NewReader(expected))
	streamer.Next()
	p := streamer.Packet()

	buf := bytes.NewBuffer(make([]byte, 4096))
	buf.Reset()

	writer := NewWriter(buf)
	n, err := writer.Write(p)
	if err != nil {
		t.Error("unexpected error", err)
	}
	if n != int64(len(expected)) {
		t.Errorf("unexpected number of bytes written %d", n)
	}

	actual := buf.Next(len(expected))

	for i, b := range expected {
		if b != actual[i] {
			t.Errorf("mismatch at index %d: %d != %d", i, actual[i], b)
		}
	}
}
