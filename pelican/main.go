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

	// Create a unique filename for this recording
	timestamp := time.Now().Format("20060102-150405")
	outputFile := "output-" + timestamp + ".mp3"

	// FFmpeg can't execute until the server is running
	for {
		cmd := exec.Command("ffmpeg",
			"-i", "rtsp://localhost:8554/stream",
			"-t", "300", // 5 minute vid then restart
			"-vn", // skip video
			"-acodec", "libmp3lame",
			"-ab", "128k", // Audio bitrate explicitly set
			outputFile, // output file
		)
		log.Printf("Attempting to connect to stream...")

		if err := cmd.Run(); err != nil {
			log.Printf("Stream not ready, retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}
}
