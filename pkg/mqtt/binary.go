package mqtt

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
)

func Uint16(buf *bytes.Buffer) uint16 {
	return binary.BigEndian.Uint16(buf.Next(2))
}

func PutUint16(buf *bytes.Buffer, val uint16) {
	buf.WriteByte(byte(val >> 8))
	buf.WriteByte(byte(val))
}

func LengthEncodedString(buf *bytes.Buffer) (str string, err error) {
	strLen := int(Uint16(buf))
	if strLen == 0 {
		return
	}

	if buf.Len() < strLen {
		err = errors.New("buffer to short")
		return
	}

	str = string(buf.Next(strLen)) // will copy the slice
	return
}

func PutLengthEncodedString(buf *bytes.Buffer, str string) {
	strLen := len(str)

	if strLen > math.MaxUint16 {
		panic("string too large")
	}

	PutUint16(buf, uint16(strLen))
	if strLen == 0 {
		return
	}
	buf.WriteString(str)
}

func LengthEncodedField(buf *bytes.Buffer) (field []byte, err error) {
	fieldLen := int(Uint16(buf))
	if fieldLen == 0 {
		field = []byte{}
		return
	}

	if buf.Len() < fieldLen {
		err = errors.New("buffer to short")
		return
	}

	field = make([]byte, fieldLen)

	_, err = buf.Read(field)
	if err != nil {
		return
	}

	return
}

func PutLengthEncodedField(buf *bytes.Buffer, field []byte) {
	fieldLen := len(field)

	if fieldLen > math.MaxUint16 {
		panic("field too large")
	}

	PutUint16(buf, uint16(fieldLen))

	if fieldLen == 0 {
		return
	}

	buf.Write(field)
}

func VariableByteUint32(buf *bytes.Buffer) (result uint32, err error) {
	var b byte

	for read := 0; read <= 4; read++ {
		b, err = buf.ReadByte()
		if err != nil {
			return
		}

		result += uint32(b&dmask) << (7 * read)

		if b&cmask == 0 {
			break
		}
	}

	return
}

func PutVariableByteUint32(buf *bytes.Buffer, val uint32) {
	var x = val // running variable that's incrementally shifted to the right

	for x > 0 {
		b := byte(x & dmask)
		if x > 127 {
			b |= cmask // add continuation bit
		}
		buf.WriteByte(b)
		x >>= 7
	}
}
