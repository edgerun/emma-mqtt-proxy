package mqtt

import (
	"bytes"
	"testing"
)

func TestScanner_RawPacket(t *testing.T) {

	input := []byte{
		// a publish packet
		48, 10, // Header (publish)
		0, 4, // Topic length
		116, 101, 115, 116, // Topic (test)
		116, 101, 115, 116, // Payload (test),
		// and another
		48, 10, // Header (publish)
		0, 4, // Topic length
		116, 101, 115, 116, // Topic (test)
		116, 101, 115, 116, // Payload (test)
	}

	reader := bytes.NewReader(input)

	scanner := NewScanner(reader)

	scanner.Scan()
	packet1 := scanner.RawPacket()
	scanner.Scan()
	packet2 := scanner.RawPacket()

	if packet1.Header.Type != TypePublish {
		t.Error("expected packet1 to be of type PUBLISH")
	}
	if packet1.Header.Length != 10 {
		t.Error("expected packet1 to be of length 10")
	}
	if packet2.Header.Length != 10 {
		t.Error("expected packet2 to be of length 10")
	}
	if scanner.Err() != nil {
		t.Error("unexpected scanner error", scanner.Err())
	}
}
