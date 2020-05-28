package mqtt

import (
	"errors"
	"io"
)

const cmask = 0b10000000 // 0x80 masks continuation bit
const dmask = 0b01111111 // 0x7F masks integer data

func WriteFull(w io.Writer, buf []byte) (n int, err error) {
	min := len(buf)

	var nn int
	for n < min && err == nil {
		nn, err = w.Write(buf[n:])
		n += nn
	}

	if n >= min {
		err = nil
	} else if n > 0 && err == io.EOF {
		err = io.ErrUnexpectedEOF
	}

	return
}

func PutVariableInt(buf []byte, i uint32) (pos int) {
	var x = i // running variable that's incrementally shifted to the right

	for x > 0 {
		b := byte(x & dmask)
		if x > 127 {
			b |= cmask // add continuation bit
		}
		buf[pos] = b

		pos++
		x >>= 7
	}

	return
}

func WritePacketHeader(writer io.Writer, header PacketHeader) (int, error) {
	// TODO seems unnecessarily complicated
	var nw int             // how many bytes the header eventually requires
	buf := make([]byte, 5) // maximal necessary byte buffer to write the header

	// packet type and flags in first byte
	buf[0] = (header.Type << 4) & 0b11110000
	buf[0] += header.Flags & 0b00001111

	nw = PutVariableInt(buf[1:], header.Length)
	nw += 1 // the header byte that was written

	return WriteFull(writer, buf[:nw])
}

func WriteRawPacket(writer io.Writer, packet *RawPacket) (n int, err error) {
	header := packet.Header

	remainingLength := int(header.Length)
	if remainingLength != len(packet.Payload) {
		return 0, errors.New("packet length mismatch")
	}

	n, err = WritePacketHeader(writer, header)
	if err != nil {
		return
	}

	np, err := WriteFull(writer, packet.Payload)
	n += np
	if err != nil {
		return
	}

	return
}

func ReadPacketHeader(reader io.Reader, header *PacketHeader) (int, error) {
	buf := []byte{0x0}

	n, err := reader.Read(buf)

	if err != nil {
		return n, err
	}
	if n == 0 {
		return n, nil
	}

	header.Type = (buf[0] >> 4) & 0xF
	header.Flags = buf[0] & 0xF

	length, nLength, err := ReadVariableByteUint32(reader)
	header.Length = length
	return n + nLength, err
}

// Reads from an io.Reader a uint32 encoded as variable byte integer in a maximum of 4 sequential bytes.
// See https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901011
func ReadVariableByteUint32(reader io.Reader) (uint32, int, error) {
	var read int
	var result uint32

	buf := []byte{0x0}
	for read = 0; read <= 4; read++ {
		n, err := reader.Read(buf)
		if err != nil {
			return 0, read + n, err
		}
		if n != 1 {
			continue
		}

		b := buf[0]
		result += uint32(b&dmask) << (7 * read)

		if b&cmask == 0 {
			break
		}
	}

	return result, read + 1, nil
}
