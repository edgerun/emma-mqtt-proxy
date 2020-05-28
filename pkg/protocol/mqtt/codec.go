package mqtt

import (
	"io"
)

func ReadPacketHeader(header *PacketHeader, reader io.Reader) (int, error) {
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

	length, nLength, err := ReadVariableByteUint32(reader) // todo blocking busy wait
	header.RemainingLength = length
	return n + nLength, err
}

// Reads from an io.Reader a uint32 encoded as variable byte integer in a maximum of 4 sequential bytes.
// See https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901011
func ReadVariableByteUint32(reader io.Reader) (uint32, int, error) {
	const cmask = 0b10000000 // 0x80 masks continuation bit
	const dmask = 0b01111111 // 0x7F masks integer data

	var read int
	var result uint32

	buf := []byte{0x0}
	for read = 0; read <= 4; read++ {
		n, err := reader.Read(buf)
		if err != nil {
			return 0, read + n, err
		}
		if n != 1 {
			continue // TODO: busy wait
		}

		b := buf[0]
		result += uint32(b&dmask) << (7 * read)

		if b&cmask == 0 {
			break
		}
	}

	return result, read + 1, nil
}
