package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
)

func main() {

	var numOfStreams = 10

	for i := 0; i < numOfStreams; i++ {
		go streamFiles(i)
	}

	// Wait for a signal to terminate all streams
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("Received shutdown signal, terminating all streams...")
}

func streamFiles(deviceNumber int) error {

	cmd := exec.Command(
		"ffmpeg",
		"-re",
		"-stream_loop", "-1",
		"-i", "./vids/wendypanel.mp4",
		"-c", "copy",
		"-f", "rtsp",
		"-rtsp_transport", "tcp",
		"rtsp://pelican:8554/stream/camera"+strconv.Itoa(deviceNumber),
	)

	cmd.Stderr = os.Stderr

	// start the stream
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("Failed to start stream: %v\n", err)
	}

	// graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan

	// kill the FFmpeg process gracefully
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("Failed to kill FFmpeg process: %v\n", err)
	}

	return nil
}
