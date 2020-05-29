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

	read, err := ReadPacketHeader(reader, header)

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

func TestReadPacketHeader_Connect(t *testing.T) {
	reader := bytes.NewReader([]byte{
		16, 31,
		0, 6, 77, 81, 73, 115, 100, 112,
		3, 2, 0, 60, 0, 17, 109, 111, 115,
		113, 112, 117, 98, 124, 54, 56, 53, 52,
		45, 116, 121, 116, 111,
	})

	header := &PacketHeader{}
	_, err := ReadPacketHeader(reader, header)

	if err != nil {
		t.Error("unexpected error", err)
	}
	if header.Type != TypeConnect {
		t.Errorf("Expected packet type %s, got %s", "CONNECT", PacketTypeName(header.Type))
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
