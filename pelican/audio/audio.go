package audio

import (
	"bytes"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/description"
	"github.com/bluenviron/gortsplib/v4/pkg/format"
	"github.com/go-audio/wav"
	"github.com/pion/rtp"
)

type Config struct {
	ListenAddress string
	PathName      string
	OutputFile    string
}

type AudioProcessor struct {
	config     Config
	server     *gortsplib.Server
	wavWriter  *wav.Writer
	audioCodev *rtspcodecs.Audio
}

type AudioHandler struct {
	buffer        bytes.Buffer
	bufferQueu    *[]bytes.Buffer // will use once we have more data to aggregate, and buffer could pop
	bufferLock    sync.Mutex
	format        format.Format
	maxBufferSize int64
	flushInterval time.Duration
	lastFlush     time.Time
}

func NewAudioHandler() *AudioHandler {

	return &AudioHandler{
		maxBufferSize: 10485760,
	}
}

// HandlePacket processes incoming RTP packets
func (h *AudioHandler) HandlePacket(media *description.Media, forma format.Format, pkt *rtp.Packet) {
	h.bufferLock.Lock()
	defer h.bufferLock.Unlock()
	if h.format == nil {
		// Store format information for conversion
		h.format = forma
	}

	// Check buffer size before writing
	if h.buffer.Len() >= int(h.maxBufferSize) {
		// Handle buffer full condition
		return
	}

	h.buffer.Write(pkt.Payload)
}

// ConvertToWav converts buffered audio data to WAV format using FFmpeg
func (h *AudioHandler) ConvertToWav(outputPath string) error {
	// Create FFmpeg command with appropriate input format
	cmd := exec.Command("ffmpeg",
		"-f", h.getFFmpegFormat(), // Input format
		"-i", "pipe:0", // Read from stdin
		"-acodec", "pcm_s16le", // Output codec (standard for WAV)
		"-ar", "16000", // Sample rate required by AWS Transcribe
		"-ac", "1", // Mono audio
		"-y",       // Overwrite output file
		outputPath, // Output file
	)

	// Create pipe to FFmpeg's stdin
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %v", err)
	}

	// Start FFmpeg process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start FFmpeg: %v", err)
	}

	// Write buffered audio data to FFmpeg
	h.bufferLock.Lock()
	_, err = stdin.Write(h.buffer.Bytes())
	h.bufferLock.Unlock()

	if err != nil {
		return fmt.Errorf("failed to write to FFmpeg: %v", err)
	}

	// Close stdin to signal EOF to FFmpeg
	stdin.Close()

	// Wait for FFmpeg to finish
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("FFmpeg failed: %v", err)
	}

	return nil
}

// getFFmpegFormat returns the appropriate FFmpeg input format based on the RTP format
func (h *AudioHandler) getFFmpegFormat() string {
	// You'll need to map your RTP format to FFmpeg format
	// This is a basic example - expand based on your needs
	switch h.format.Codec() {
	case "PCMA":
		return "alaw"
	case "PCMU":
		return "mulaw"
	case "L16":
		return "s16be"
	default:
		return "s16le" // default format
	}
}
