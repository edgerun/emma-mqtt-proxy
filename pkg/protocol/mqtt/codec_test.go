package mqtt

import (
	"bytes"
	"io"
	"testing"
)

func TestReadVariableByteUint32(t *testing.T) {
	// TODO: write more tests (+ error cases)

	matrix := []struct{
		reader io.Reader
		expected uint32
	}{
		{
			bytes.NewReader([]byte{
				0b10000000, // cont
				0b00000001, // 128
			}), 128,
		},
		{
			bytes.NewReader([]byte{
				0b00000001, // cont
			}), 1,
		},
	}

	for _, testcase := range matrix {
		actual, err := ReadVariableByteUint32(testcase.reader)

		if err != nil {
			t.Error("Unexpected error", err)
		}
		if actual != testcase.expected {
			t.Errorf("Expected val to be %d, was %d", testcase.expected, actual)
		}
	}

}
