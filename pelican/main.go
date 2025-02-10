package main

import (
	"fmt"
	"log"
	"muse/pelican/cloud"
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
			fmt.Printf("start and wait error: %v\n", err)
			log.Fatal()
		}
	}()

	// FFmpeg can't execute until the server is running
	for {

		// Create a unique filename for this recording
		timestamp := time.Now().Format("20060102-150405")
		outputFileName := "output-" + timestamp + ".mp3"

		cmd := exec.Command("ffmpeg",
			"-i", "rtsp://localhost:8554/stream",
			"-t", "30", // 30sec stream then restart
			"-vn", // cut out video, only audio added to file
			"-acodec", "libmp3lame",
			"-ab", "128k", // Audio bitrate explicitly set
			outputFileName, // output file
		)

		if err := cmd.Run(); err != nil {
			log.Printf("Stream not ready, retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
			continue
		}

		// possibly rewrite this as a go routine, so it doesn't hold up the next recording...
		err := cloud.UploadFileToS3(outputFileName)
		if err != nil {
			fmt.Printf("error trying to upload to s3: %v\n", err)
		}
		fmt.Println("S3 file upload complete...")
	}
}
