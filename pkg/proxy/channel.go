package proxy

import (
	"github.com/edgerun/emma-mqtt-proxy/pkg/mqtt"
	"log"
	"net"
)

type ReaderState int
type WriterState int

const (
	ReaderStateUnknown ReaderState = iota
	ReaderStateRunning
	ReaderStateStopped
	ReaderStateErrored
)

const (
	WriterStateUnknown WriterState = iota
	WriterStateRunning
	WriterStateStopped
	WriterStateErrored
)

type Channel struct {
	conn           net.Conn
	outbound       chan mqtt.Packet
	inbound        chan mqtt.Packet
	errors         chan error
	packetListener chan mqtt.Packet

	readerState ReaderState
	writerState WriterState

	// internal communication channels
	commConn       chan net.Conn
	commStop       chan bool
	commStopWriter chan bool
	commError      chan error
}

func NewChannel() *Channel {
	return &Channel{
		outbound: make(chan mqtt.Packet),
		inbound:  make(chan mqtt.Packet),
		errors:   make(chan error),

		readerState: ReaderStateUnknown,
		writerState: WriterStateUnknown,

		commConn:       make(chan net.Conn, 1),
		commStop:       make(chan bool),
		commStopWriter: make(chan bool),
		commError:      make(chan error),
	}
}

func (c *Channel) SetPacketListener(ch chan mqtt.Packet) {
	c.packetListener = ch
}

func (c *Channel) SetConnection(conn net.Conn) {
	c.commConn <- conn
}

func (c *Channel) Run() error {
	for {
		select {
		case err := <-c.errors:
			if err != nil {
				return err // FIXME
			}
		case conn := <-c.commConn:
			// TODO: stop current reader
			c.conn = conn
			if c.readerState != ReaderStateRunning {
				log.Println("starting read loop")
				go c.readLoop(conn)
			}
			if c.writerState != WriterStateRunning {
				log.Println("starting read write")
				go c.writeLoop(conn)
			}
		case stop := <-c.commStop:
			if stop {
				return nil
			}
		}
	}
}

func (c *Channel) Stop() {
	c.commStop <- true
}

func (c *Channel) ReaderState() ReaderState {
	return c.readerState
}

func (c *Channel) WriterState() WriterState {
	return c.writerState
}

func (c *Channel) readLoop(conn net.Conn) {
	streamer := mqtt.NewStreamer(conn)

	log.Printf("reading packets from [%s]...\n", conn.RemoteAddr())
	c.readerState = ReaderStateRunning
	for streamer.Next() {
		if streamer.Err() != nil {
			log.Printf("[%s] error while reading packet stream: %s\n", conn.RemoteAddr(), streamer.Err())
			break
		}

		packet := streamer.Packet()
		log.Printf("[%s] received packet: %s\n", conn.RemoteAddr(), mqtt.PacketTypeName(packet.Type()))
		c.packetListener <- packet

		log.Println("waiting for next packet ...")
	}

	if streamer.Err() != nil {
		log.Println("error while reading packet stream", streamer.Err())
		c.readerState = ReaderStateErrored
		c.errors <- streamer.Err()
	} else {
		c.readerState = ReaderStateStopped
	}

	log.Println("returning from packet streamer")
}

func (c *Channel) Send(packet mqtt.Packet) {
	c.outbound <- packet
}

func (c *Channel) currentConn() net.Conn {
	// TODO: mutex with setting new conn
	return c.conn
}

func (c *Channel) writeLoop(conn net.Conn) {
	w := mqtt.NewWriter(conn)
	var n int64
	var err error

	c.writerState = WriterStateRunning
	for packet := range c.outbound {
		// TODO: maybe it's a bit constricting if we have only a single write loop per design if a channel goes to a
		//  broker, we may have frequent writes which may turn out not to scale. on the other hand, the synchronization
		//  in making concurrent packet writes mutually exclusive seems to involve a lot of synchronization complexity.
		n, err = w.Write(packet)
		log.Printf("wrote %d bytes to [%s]\n", n, conn.RemoteAddr())

		if err != nil {
			c.writerState = WriterStateErrored
			c.errors <- err
			return
		}
	}
	c.writerState = WriterStateStopped
	return
}
