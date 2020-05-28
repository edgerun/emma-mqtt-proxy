package mqtt

import (
	"bytes"
	"io"
	"testing"
)

func TestReadPacketHeader(t *testing.T) {
	header := &PacketHeader{}

	reader := bytes.NewReader([]byte{
		0b11000000, // PINGREQ + no flags)
		0b00000000, // remaining length 0,
		0b11010000, // ... some next packet
	})

	read, err := ReadPacketHeader(header, reader)

	if err != nil {
		t.Error("Unexpected error", err)
	}
	if header.Type != TypePingReq {
		t.Errorf("Expected packet type %s, got %s", "PINGREQ", PacketTypeName(header.Type))
	}
	if read != 2 {
		t.Error("Expected to read 2 bytes from reader")
	}

}

func TestReadVariableByteUint32(t *testing.T) {
	// TODO: write more tests (+ error cases)

	matrix := []struct {
		reader       io.Reader
		expectedRead int
		expected     uint32
	}{
		{
			bytes.NewReader([]byte{
				0b10000000, // cont
				0b00000001, // 128
			}), 2, 128,
		},
		{
			bytes.NewReader([]byte{
				0b00000001, // cont
			}), 1, 1,
		},
	}

	for _, testcase := range matrix {
		actual, read, err := ReadVariableByteUint32(testcase.reader)

		if err != nil {
			t.Error("Unexpected error", err)
		}
		if actual != testcase.expected {
			t.Errorf("Expected val to be %d, was %d", testcase.expected, actual)
		}
		if read != testcase.expectedRead {
			t.Errorf("Expected to read %d bytes, was %d", testcase.expectedRead, read)
		}
	}

}
