package main

import (
	"log"
	"pelican/server"

	"github.com/bluenviron/gortsplib/v4"
)

func main() {
	// Create the stream manager
	streamMgr := server.NewStreamManager()

	// Create the server handler with the stream manager
	h := server.NewServerHandler(streamMgr)

	// Configure the server
	h.S = &gortsplib.Server{
		Handler:           h,
		RTSPAddress:       ":8554",
		UDPRTPAddress:     ":8000",
		UDPRTCPAddress:    ":8001",
		MulticastIPRange:  "224.1.0.0/16",
		MulticastRTPPort:  8002,
		MulticastRTCPPort: 8003,
	}

	// Start server
	log.Printf("RTSP server starting on port 8554")
	if err := h.S.StartAndWait(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
