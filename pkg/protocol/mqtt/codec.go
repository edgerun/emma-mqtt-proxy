package mqtt

import (
	"io"
)

// Reads from an io.Reader a uint32 encoded as variable byte integer in a maximum of 4 sequential bytes.
// See https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901011
func ReadVariableByteUint32(reader io.Reader) (uint32, error) {
	const cmask = 0b10000000 // 0x80 masks continuation bit
	const dmask = 0b01111111 // 0x7F masks integer data

	var pos uint8
	var result uint32

	buf := []byte{0x0}
	for pos = 0; pos <= 4; pos++ {
		n, err := reader.Read(buf)
		if err != nil {
			return 0, err
		}
		if n != 1 {
			continue // TODO: busy wait
		}

		b := buf[0]
		result += uint32(b & dmask) << (7 * pos)

		if b & cmask == 0 {
			break
		}
	}

	return result, nil
}
