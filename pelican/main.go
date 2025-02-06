package main

import (
	"log"
	"muse/pelican/server"
	"os/exec"
	"time"

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

	// Start server in a goroutine
	go func() {
		log.Printf("server is ready")
		if err := h.S.StartAndWait(); err != nil {
			log.Fatal(err)
		}
	}()

	// Wait and retry FFmpeg until stream is available
	for {
		cmd := exec.Command("ffmpeg", "-i", "rtsp://localhost:8554/stream", "-vn", "-acodec", "libmp3lame", "output.mp3")
		log.Printf("Attempting to connect to stream...")

		if err := cmd.Run(); err != nil {
			log.Printf("Stream not ready, retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}

}
