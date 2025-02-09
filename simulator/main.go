package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func main() {
	cmd := exec.Command(
		"ffmpeg",
		"-re",
		"-stream_loop", "-1",
		"-i", "./vids/wendypanel.mp4",
		"-c", "copy",
		"-f", "rtsp",
		"-rtsp_transport", "tcp",
		"rtsp://localhost:8554/stream",
	)

	cmd.Stderr = os.Stderr

	// start the stream
	if err := cmd.Start(); err != nil {
		log.Fatal("Failed to start stream:", err)
	}

	fmt.Println("Stream started at rtsp://localhost:8554/stream")
	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan

	// kill the FFmpeg process gracefully
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		log.Printf("Failed to kill FFmpeg process: %v", err)
	}

	fmt.Println("\nStream stopped")
}
