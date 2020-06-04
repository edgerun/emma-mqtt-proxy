package mqtt

import (
	"io"
)

type dropFilter struct {
}

func (filter *dropFilter) onHeader(header *PacketHeader) FilterAction {
	return ActionDropPacket
}

var defaultFilter *dropFilter = &dropFilter{}

type Streamer struct {
	r      io.Reader     // the underlying stream
	header *PacketHeader // current packet header
	err    error         // the first non-eof error that occurred
	done   bool          // if we're done

	filter   HeaderFilter
	handlers map[FilterAction]PacketHandler
}

func NewStreamer(r io.Reader) *Streamer {
	s := &Streamer{
		r:        r,
		filter:   defaultFilter,
		handlers: make(map[FilterAction]PacketHandler),
	}

	s.handlers[ActionDropPacket] = NewPacketDropper()

	return s
}

func (s *Streamer) SetFilter(filter HeaderFilter) {
	s.filter = filter
}

func (s *Streamer) SetHandler(action FilterAction, handler PacketHandler) {
	s.handlers[action] = handler
}

func (s *Streamer) Err() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}

func (s *Streamer) setErr(err error) {
	if s.err == nil || s.err == io.EOF {
		s.err = err
	}
}

func (s *Streamer) bail(err error) bool {
	s.setErr(err)
	s.done = true
	return false
}

func (s *Streamer) Next() bool {
	if s.done {
		return false
	}

	r := s.r

	// read the packet header
	header, err := ReadHeaderFrom(r)
	if err != nil {
		return s.bail(err)
	}

	// TODO: apply filters and potentially drop packet, or skip decoding and copy data directly

	action := s.filter.onHeader(header)

	handler, ok := s.handlers[action]
	if !ok {
		// TODO: discard packet
	} else {
		handler.onPacket(header, r)
	}

	// read the packet
	s.packet, err = s.readPacket(header)
	if err != nil {
		return s.bail(err)
	}

	return true
}
