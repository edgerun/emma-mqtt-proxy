package mqtt

import "testing"

func TestPacketHeader_ByteSize(t *testing.T) {
	var header PacketHeader

	header = PacketHeader{0, 0, 0}
	if header.byteSize() != 2 {
		t.Errorf("Expected a 0 length packet to have a header of size 2, not %d", header.byteSize())
	}

	header = PacketHeader{0, 0, 127}
	if header.byteSize() != 2 {
		t.Errorf("Expected a 127 length packet to have a header of size 2, not %d", header.byteSize())
	}

	header = PacketHeader{0, 0, 128}
	if header.byteSize() != 3 {
		t.Errorf("Expected a 128 length packet to have a header of size 3, not %d", header.byteSize())
	}

	// 1|0000000
	// 0|1000000 => 10000000000000
	header = PacketHeader{0, 0, 8192}
	if header.byteSize() != 3 {
		t.Errorf("Expected a 8192 length packet to have a header of size 3, not %d", header.byteSize())
	}

	// 1|0000000
	// 1|0000000
	// 0|0000001 => 100000000000000 = 16384
	header = PacketHeader{0, 0, 16384}
	if header.byteSize() != 4 {
		t.Errorf("Expected a 16384 length packet to have a header of size 4, not %d", header.byteSize())
	}

	// 1|1111111
	// 0|1111111 => 11111111111111
	header = PacketHeader{0, 0, 16383}
	if header.byteSize() != 3 {
		t.Errorf("Expected a 16383 length packet to have a header of size 3, not %d", header.byteSize())
	}

}

