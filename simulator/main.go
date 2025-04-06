package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

func main() {
	var numOfStreams = 10
	var wg sync.WaitGroup

	// Create a channel to signal all goroutines to terminate
	stopChan := make(chan struct{})

	// Start all streams
	for i := 0; i < numOfStreams; i++ {
		wg.Add(1)
		go func(deviceNum int) {
			defer wg.Done()
			err := streamFiles("camera", deviceNum, stopChan)
			if err != nil {
				fmt.Printf("Stream %d error: %v\n", deviceNum, err)
			}
		}(i)
	}

	// Wait for a signal to terminate all streams
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("Received shutdown signal, terminating all streams...")
	close(stopChan) // Signal all goroutines to stop

	// Wait for all streams to shut down gracefully
	wg.Wait()
	fmt.Println("All streams terminated successfully")
}

func streamFiles(deviceType string, deviceNumber int, stopChan <-chan struct{}) error {
	streamURL := "rtsp://pelican:8554/stream/" + deviceType + strconv.Itoa(deviceNumber)

	fmt.Printf("Starting stream %d to %s\n", deviceNumber, streamURL)

	cmd := exec.Command(
		"ffmpeg",
		"-re",
		"-stream_loop", "-1",
		"-i", "./vids/wendypanel.mp4",
		"-c", "copy",
		"-f", "rtsp",
		"-rtsp_transport", "tcp",
		streamURL,
	)

	// Capture stderr to a pipe so we can detect when the stream is properly started
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	// Start the stream
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start stream: %v", err)
	}

	// Read output from FFmpeg to confirm it's working
	go func() {
		buffer := make([]byte, 4096)
		confirmationReceived := false

		for {
			n, err := stderr.Read(buffer)
			if err != nil {
				break
			}

			output := string(buffer[:n])

			// Only print FFmpeg output if there's an error
			if !confirmationReceived && isStreamStarted(output) {
				fmt.Printf("✅ Stream %d confirmed running\n", deviceNumber)
				confirmationReceived = true
			}
		}
	}()

	// Print a periodic heartbeat to confirm the stream is still running
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Wait for either stop signal or process completion
	go func() {
		select {
		case <-stopChan:
			// Kill the FFmpeg process gracefully when stop signal received
			if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
				fmt.Printf("Failed to kill FFmpeg process %d: %v\n", deviceNumber, err)
			}
		case <-ticker.C:
			fmt.Printf("Stream %d still running\n", deviceNumber)
		}
	}()

	// Wait for the process to finish
	err = cmd.Wait()
	if err != nil {
		// This is expected when we terminate the process
		if _, ok := err.(*exec.ExitError); ok {
			fmt.Printf("Stream %d terminated\n", deviceNumber)
			return nil
		}
		return fmt.Errorf("stream process error: %v", err)
	}

	return nil
}

// isStreamStarted checks FFmpeg output to see if streaming has begun
func isStreamStarted(output string) bool {
	// This string appears in FFmpeg output when an RTSP stream is established
	return contains(output, "Stream mapping:") ||
		contains(output, "rtsp://") ||
		contains(output, "Output #0")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
