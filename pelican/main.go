package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"muse/pelican/audio"
	pb "muse/pelican/pkg/file"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/base"
	"github.com/bluenviron/gortsplib/v4/pkg/description"
	"github.com/pion/rtp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const chunkSize = 64 * 1024 // 64KB chunks

type GRPCClient struct {
	conn   *grpc.ClientConn
	client pb.AudioFileServiceClient
}

func NewGRPCClient(serverAddr string) (*GRPCClient, error) {
	// Create a connection to the server
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	client := pb.NewAudioFileServiceClient(conn)
	return &GRPCClient{
		conn:   conn,
		client: client,
	}, nil
}

func (c *GRPCClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *GRPCClient) UploadWavFile(ctx context.Context, filePath string) (*pb.UploadStatus, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Get file info for total size
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %v", err)
	}

	// Create upload stream
	stream, err := c.client.UploadFile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload stream: %v", err)
	}

	// Prepare buffer for chunks
	buffer := make([]byte, chunkSize)
	chunkIndex := int32(0)

	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading file: %v", err)
		}

		chunk := &pb.FileChunk{
			ContentType: "audio/wav",
			ChunkData:   buffer[:n],
			ChunkIndex:  chunkIndex,
			TotalSize:   fileInfo.Size(),
		}

		// Set additional metadata in the first chunk
		if chunkIndex == 0 {
			chunk.FileName = filepath.Base(filePath)
		}

		// Set isLast flag if this is the final chunk
		if n < chunkSize {
			chunk.IsLast = true
		}

		if err := stream.Send(chunk); err != nil {
			return nil, fmt.Errorf("failed to send chunk: %v", err)
		}

		chunkIndex++
	}

	// Close the stream and get the response
	status, err := stream.CloseAndRecv()
	if err != nil {
		return nil, fmt.Errorf("failed to receive upload status: %v", err)
	}

	return status, nil
}

func (c *GRPCClient) HealthCheck(ctx context.Context) (*pb.HealthCheckResponse, error) {
	return c.client.HealthCheck(ctx, &pb.HealthCheckRequest{})
}

func main() {
	// Configuration
	config := struct {
		rtspURL        string
		grpcServerAddr string
		recordInterval time.Duration
	}{
		rtspURL:        "rtsp://localhost:8554/stream",
		grpcServerAddr: "localhost:50051",
		recordInterval: 30 * time.Second,
	}

	// Initialize GRPC client
	grpcClient, err := NewGRPCClient(config.grpcServerAddr)
	if err != nil {
		log.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer grpcClient.Close()

	// Create RTSP client
	rtspClient := gortsplib.Client{}

	// Parse RTSP URL
	url, err := base.ParseURL(config.rtspURL)
	if err != nil {
		log.Fatalf("Failed to parse RTSP URL: %v", err)
	}

	// Connect to RTSP server
	err = rtspClient.Start(url.Scheme, url.Host)
	if err != nil {
		log.Fatalf("Failed to connect to RTSP server: %v", err)
	}
	defer rtspClient.Close()

	// Get stream description
	desc, _, err := rtspClient.Describe(url)
	if err != nil {
		log.Fatalf("Failed to get stream description: %v", err)
	}

	// Find and setup audio media
	var audioMedia *description.Media
	for _, media := range desc.Medias {
		if media.Type == "audio" {
			audioMedia = media
			break
		}
	}

	if audioMedia == nil {
		log.Fatal("No audio stream found")
	}

	// Setup the audio stream
	_, err = rtspClient.Setup(desc.BaseURL, audioMedia, 0, 0)
	if err != nil {
		log.Fatalf("Failed to setup audio stream: %v", err)
	}

	// Initialize audio handler
	audioHandler := audio.NewAudioHandler()

	// Start stream
	_, err = rtspClient.Play(nil)
	if err != nil {
		log.Fatalf("Failed to start playing: %v", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle audio packets
	rtspClient.OnPacketRTP(audioMedia, audioMedia.Formats[0], func(pkt *rtp.Packet) {
		audioHandler.HandlePacket(audioMedia, audioMedia.Formats[0], pkt)
	})

	// Start periodic recording and upload
	go func() {
		ticker := time.NewTicker(config.recordInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				wavFile := fmt.Sprintf("audio_%d.wav", time.Now().Unix())

				err := audioHandler.ConvertToWav(wavFile)
				if err != nil {
					log.Printf("Failed to convert to WAV: %v", err)
					continue
				}

				status, err := grpcClient.UploadWavFile(ctx, wavFile)
				if err != nil {
					log.Printf("Failed to upload file: %v", err)
				} else {
					log.Printf("Successfully uploaded file: %s", status.S3Path)
				}

				os.Remove(wavFile)
			}
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down gracefully...")
}
