package mqtt

import (
	"bytes"
	"io"
)

type Writer struct {
	w    io.Writer     // the underlying writer to write to
	hBuf *bytes.Buffer // buffer used for the header
	pBuf *bytes.Buffer // buffer used for the packet
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w:    w,
		hBuf: bytes.NewBuffer(make([]byte, 5)),
		pBuf: bytes.NewBuffer(make([]byte, 4096)),
	}
}

// Write the packet into the underlying io.Writer. It does this as follows:
//  1. serialize the packet into a byte buffer to know how long it is
//  2. update the remaining length field of the header to the length that was written into byte buffer holding the packet
//  3. serialize the header into a byte buffer
//  4. write the header buffer
//  5. write the packet buffer
func (w *Writer) Write(packet Packet) (written int64, err error) {
	// reset buffers
	hBuf := w.hBuf
	pBuf := w.pBuf
	hBuf.Reset()
	pBuf.Reset()

	// write packet into packet buffer
	err = EncodePacket(pBuf, packet)
	if err != nil {
		return
	}

	// create a new header for the outgoing packet that has as remaining length the length of the encoded packet (may be
	// different from the original incoming packet)
	h := &PacketHeader{
		Length: uint32(pBuf.Len()),
		Type: packet.Type(),
		Flags: packet.Flags(),
	}
	packet.setHeader(h) // does it really make sense to change the header of the original packet?

	// write header into header buffer
	err = EncodeHeader(hBuf, h)
	if err != nil {
		return
	}

	// write header buffer
	n, err := hBuf.WriteTo(w.w)
	written += n
	if err != nil {
		return
	}

	// write packet buffer
	n, err = pBuf.WriteTo(w.w)
	written += n
	if err != nil {
		return
	}

	return
}
