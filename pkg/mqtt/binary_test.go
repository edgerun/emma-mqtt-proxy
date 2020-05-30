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

func TestPutLengthEncodedString(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 32))
	buf.Reset()

	PutLengthEncodedString(buf, "MQTT")

	actual := buf.Next(6)
	expected := []byte{
		0b00000000, // 0 (Length MSB)
		0b00000100, // 4 (Length LSB)
		0b01001101, // M
		0b01010001, // Q
		0b01010100, // T
		0b01010100, // T
	}

	for i, b := range actual {
		if expected[i] != b {
			t.Errorf("Mismatch at index %d: %b != %b", i, b, expected[i])
		}
	}
}

func TestLengthEncodedStringIntegration(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 32))
	buf.Reset()

	str := "MQTT"

	PutLengthEncodedString(buf, str)
	actual, err := LengthEncodedString(buf)
	if err != nil {
		t.Error("unexpected error", err)
	}

	if actual != str {
		t.Errorf("mismatch %s != %s", str, actual)
	}

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

func TestLengthEncodedString_WithNumbers(t *testing.T) {
	slice := []byte{
		0, 15, // length
		109, 111, 115, 113, 112, 117, 98, // mosqpub
		124, 57, 52, 48, 56, 45, 111, 109, // |9408-om
	}

	buf := bytes.NewBuffer(make([]byte, 512))
	buf.Reset()
	buf.Write(slice)

	str, err := LengthEncodedString(buf)

	if err != nil {
		t.Error("unexpected error", err)
	}

	if str != "mosqpub|9408-om" {
		t.Error("")
	}

}

func TestPutLengthEncodedString_WithNumbers(t *testing.T) {
	expected := []byte{
		0, 15, // length
		109, 111, 115, 113, 112, 117, 98, // mosqpub
		124, 57, 52, 48, 56, 45, 111, 109, // |9408-om
	}

	buf := bytes.NewBuffer(make([]byte, 512))
	buf.Reset()

	PutLengthEncodedString(buf, "mosqpub|9408-om")

	if buf.Len() != 17 {
		t.Error("expected 17 bytes to be written, was", buf.Len())
	}

	actual := buf.Next(17)

	for i, b := range actual {
		if b != expected[i] {
			t.Errorf("error at index %d: %d != %d", i, b, expected[i])
		}
	}
}
