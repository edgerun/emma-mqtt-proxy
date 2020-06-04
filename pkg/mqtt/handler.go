package mqtt

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
)

type PacketHandler interface {
	onPacket(header *PacketHeader, r io.Reader) error
}

type PacketForwarder struct {
	limR *io.LimitedReader
	W    io.Writer // the writer to forward to
}

type PacketDropper struct {
	limR *io.LimitedReader
}

type PacketDecoder struct {
	packet Packet
	limR   *io.LimitedReader // limited reader for reading packet payloads into buffer from stream
	buf    *bytes.Buffer     // the buffer that holds packet data to be unmarshalled
}

type PacketDecoderChannel struct {
	decoder *PacketDecoder
	channel chan<- Packet
}

func NewPacketForwarder() *PacketForwarder {
	return &PacketForwarder{}
}

func NewPacketDropper() *PacketDropper {
	return &PacketDropper{
		limR: &io.LimitedReader{},
	}
}

func NewPacketDecoder() *PacketDecoder {
	return &PacketDecoder{
		limR: &io.LimitedReader{},
	}
}

func NewPacketDecoderChannel(channel chan<- Packet) *PacketDecoderChannel {
	return &PacketDecoderChannel{
		decoder: NewPacketDecoder(),
		channel: channel,
	}
}

func (handler *PacketDropper) onPacket(header *PacketHeader, r io.Reader) (err error) {
	handler.limR.R = r
	handler.limR.N = int64(header.Length)

	_, err = io.Copy(ioutil.Discard, handler.limR)

	return
}

func (handler *PacketForwarder) onPacket(header *PacketHeader, r io.Reader) (err error) {
	handler.limR.R = r
	handler.limR.N = int64(header.Length)

	_, err = io.Copy(handler.W, handler.limR)

	return
}

func (handler *PacketDecoder) onPacket(header *PacketHeader, r io.Reader) (err error) {
	var p Packet
	// prepare limited reader to read the packet length exactly from the underlying reader
	buf := handler.buf

	if buf.Len() != 0 {
		panic(fmt.Sprintf("packet buffer should be empty, was %d", buf.Len()))
	}
	handler.limR.R = r
	handler.limR.N = int64(header.Length)

	// prepare buffer to read into
	buf.Reset()
	buf.Grow(int(header.Length)) // make sure we have enough space

	// read packet data into buffer
	_, err = buf.ReadFrom(handler.limR)
	if err != nil {
		return
	}

	// unmarshal packet into buffer (we've ensured that s.buf is exactly the remaining length of the MQTT packet)
	p, err = DecodePacket(buf, header)
	if err != nil {
		return
	}

	handler.packet = p
	return
}

func (handler *PacketDecoderChannel) onPacket(header *PacketHeader, r io.Reader) (err error) {
	err = handler.decoder.onPacket(header, r)
	if err != nil {
		return
	}

	handler.channel <- handler.decoder.packet

	return
}
