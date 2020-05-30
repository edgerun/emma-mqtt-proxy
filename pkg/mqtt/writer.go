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

func (w *Writer) Write(packet Packet) (written int64, err error) {
	// reset buffers
	hBuf := w.hBuf
	pBuf := w.pBuf
	hBuf.Reset()
	pBuf.Reset()

	// write packet into packet buffer
	err = EncodePacket(pBuf, packet, packet.Type())
	if err != nil {
		return
	}

	// set remaining length in header to length of the encoded packet (may be different from the original)
	packet.Header().Length = uint32(pBuf.Len())

	// write header into header buffer
	err = EncodeHeader(hBuf, packet.Header())
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
