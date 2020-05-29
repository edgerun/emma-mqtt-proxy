package mqtt

import (
	"bytes"
	"io"
	"testing"
)

func TestLengthEncodedString(t *testing.T) {
	byteArray := []byte{
		0b00000000, // 0 (Length MSB)
		0b00000100, // 4 (Length LSB)
		0b01001101, // M
		0b01010001, // Q
		0b01010100, // T
		0b01010100, // T
	}

	buf := bytes.NewBuffer(byteArray)

	protocolName, err := LengthEncodedString(buf)

	if err != nil {
		t.Error("unexpected error:", err)
	}
	if protocolName != "MQTT" {
		t.Error("unepxected value:", protocolName)
	}

	println(buf.Len())
	println(string(buf.Next(2)))
}

func TestVariableByteUint32(t *testing.T) {
	// TODO: write more tests (+ error cases)

	matrix := []struct {
		reader       io.Reader
		expectedRead int64
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
		buf := bytes.NewBuffer(make([]byte, 1024))
		buf.Reset()

		r, err := buf.ReadFrom(testcase.reader)
		if err != nil {
			t.Error("unexpected error", err)
		}
		if r != testcase.expectedRead {
			t.Error("Did not read expected number of bytes", r)
		}

		actual, err := VariableByteUint32(buf)

		if err != nil {
			t.Error("Unexpected error", err)
		}
		if actual != testcase.expected {
			t.Errorf("Expected val to be %d, was %d", testcase.expected, actual)
		}
	}
}


func TestName(t *testing.T) {
	slice := []byte{
		0b00000000, // 0 (Length MSB)
		0b00000100, // 4 (Length LSB)
		0b01001101, // M
		0b01010001, // Q
		0b01010100, // T
		0b01010100, // T
		0b01010100, // T
		0b01010001, // Q
	}

	reader := bytes.NewReader(slice)

	buf := bytes.NewBuffer(make([]byte, 20))
	buf.Reset()

	lr := io.LimitReader(reader, 6)

	n, _ := buf.ReadFrom(lr)
	println("read", n, "bytes")

}
