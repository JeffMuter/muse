package main

import (
	"log"
	"muse/pelican/server"

	"github.com/bluenviron/gortsplib/v4"
)

func main() {
	// configure the server
	h := &server.ServerHandler{}
	h.S = &gortsplib.Server{
		Handler:           h,
		RTSPAddress:       ":8554",
		UDPRTPAddress:     ":8000",
		UDPRTCPAddress:    ":8001",
		MulticastIPRange:  "224.1.0.0/16",
		MulticastRTPPort:  8002,
		MulticastRTCPPort: 8003,
	}

	// start server and wait until a fatal error
	log.Printf("server is ready")
	panic(h.S.StartAndWait())

}
