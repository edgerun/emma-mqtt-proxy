package mqtt

import (
	"bytes"
	"encoding/binary"
	"errors"
)

func Uint16(buf *bytes.Buffer) uint16 {
	return binary.BigEndian.Uint16(buf.Next(2))
}

func LengthEncodedString(buf *bytes.Buffer) (str string, err error) {
	strlen := int(Uint16(buf))
	if strlen == 0 {
		return
	}

	minlen := buf.Len()

	if minlen < strlen {
		err = errors.New("buffer to short")
		return
	}

	str = string(buf.Next(strlen))
	return
}

func VariableByteUint32(buf *bytes.Buffer) (result uint32, err error) {
	var b byte

	for read := 0; read <= 4; read++ {
		b, err = buf.ReadByte()
		if err != nil {
			return
		}
		println("reading ", read, "got", b)

		result += uint32(b&dmask) << (7 * read)

		if b&cmask == 0 {
			break
		}
	}

	return
}
