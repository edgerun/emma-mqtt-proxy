package proxy

import (
	"github.com/edgerun/emma-mqtt-proxy/pkg/mqtt"
	"io"
	"log"
	"sync"
)

type Router func(header *mqtt.PacketHeader) mqtt.Writer

type Bridge struct {
	left  mqtt.Channel
	right mqtt.Channel

	lRouter Router
	rRouter Router
	lStream *RoutingStreamer
	rStream *RoutingStreamer

	wg sync.WaitGroup
}

func NewBridge(left io.ReadWriter, right io.ReadWriter) (b *Bridge) {
	return NewChannelBridge(mqtt.NewChannel(left), mqtt.NewChannel(right))
}

// Creates a new bridge between two channels. The caller yields control of the channels (in particular reading from the
// stream) to the bridge. The default routers simply forward packets to the opposite side.
func NewChannelBridge(left mqtt.Channel, right mqtt.Channel) (b *Bridge) {
	b = &Bridge{
		left:  left,
		right: right,
	}

	// default routers simply forward to the opposite MQTT channel
	b.lRouter = func(header *mqtt.PacketHeader) mqtt.Writer { return b.SinkRight() }
	b.rRouter = func(header *mqtt.PacketHeader) mqtt.Writer { return b.SinkLeft() }

	// pass wrappers to the routing streamers that
	b.lStream = NewRoutingStreamer(left, b.routeLeftToRight)
	b.rStream = NewRoutingStreamer(right, b.routeRightToLeft)

	return
}

func (b *Bridge) SinkLeft() mqtt.PacketSink {
	return b.left
}

func (b *Bridge) SinkRight() mqtt.PacketSink {
	return b.right
}

func (b *Bridge) SetRouterLeft(router Router) {
	b.lRouter = router
}

func (b *Bridge) SetRouterRight(router Router) {
	b.rRouter = router
}

func (b *Bridge) routeLeftToRight(header *mqtt.PacketHeader) mqtt.Writer {
	return b.lRouter(header)
}

func (b *Bridge) routeRightToLeft(header *mqtt.PacketHeader) mqtt.Writer {
	return b.rRouter(header)
}

func (b *Bridge) Wait() {
	b.wg.Wait()
}

func (b *Bridge) Start() chan error {
	errs := make(chan error, 2)
	b.wg.Add(2)

	go func() {
		go func() { errs <- b.lStream.Run(); b.wg.Done() }()
		go func() { errs <- b.rStream.Run(); b.wg.Done() }()

		b.wg.Wait()
		close(errs)
		log.Println("closing bridge")
	}()

	return errs
}

type RoutingStreamer struct {
	streamer mqtt.Streamer
	router   Router
}

func NewRoutingStreamer(streamer mqtt.Streamer, router Router) *RoutingStreamer {
	return &RoutingStreamer{streamer: streamer, router: router}
}

func (e *RoutingStreamer) Next() (header *mqtt.PacketHeader, err error) {
	header, err = e.streamer.Next()
	if err != nil {
		return
	}

	sink := e.router(header)
	if sink == nil {
		panic("router returned is nil")
	}
	err = mqtt.Copy(e.streamer, sink)
	return
}

func (e *RoutingStreamer) Run() error {
	for {
		_, err := e.Next()
		if err != nil {
			return err
		}
	}
}
