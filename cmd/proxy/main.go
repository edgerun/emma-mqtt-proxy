package main

import (
	"flag"
	"fmt"
	"github.com/edgerun/emma-mqtt-proxy/pkg/proxy"
)

func main() {
	hostPtr := flag.String("host", "127.0.0.1", "host to bind to")
	portPtr := flag.Int("port", 1883, "the server port")

	flag.Parse()

	address := fmt.Sprintf("%s:%d", *hostPtr, *portPtr)
	proxy.Serve("tcp", address)
}
