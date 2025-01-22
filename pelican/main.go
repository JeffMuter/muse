package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"muse/pelican/audio"
	"muse/pelican/pkg"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/base"
	"github.com/bluenviron/gortsplib/v4/pkg/description"
	"github.com/bluenviron/gortsplib/v4/pkg/format"
	"github.com/pion/rtp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type StreamProcessor struct {
	inputStreams []string
	outputAddr   string
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	processes    []*exec.Cmd
	mu           sync.Mutex
}
	
	const (
    chunkSize = 64 * 1024 // 64KB chunks
    grpcTimeout = 5 * time.Minute
)
	
	 type GRPCCl struct {
    client     pkg.AudioFileServiceClient
    conn       *grpc.ClientConn
}
	
func NewGRPCClient(serverAddr string) (*GRPCClient, error) {
    conn, err := grpc.Dial(serverAddr, 
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*20)), // 20MB
		)		
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %v", err)
	}

	return &GRPCClient{
		client: pkg.NewAudioFileServiceClient(conn),
		conn:   conn,
	}, nil
}

func (g *GRPCClient) Close() error {
	return g.conn.Close()
}

func (g *GRPCClient) UploadWavFile(ctx context.Context, filepath string) (*pkg.UploadStatus, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %v", err)
	}

	ctx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()

	stream, err := g.client.UploadFile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload stream: %v", err)
	}

	// Send first chunk with metadata
	firstChunk := &pkg.FileChunk{
		FileName:    filepath.Base(filepath),
		ContentType: "audio/wav",
		TotalSize:   fileInfo.Size(),
		ChunkIndex:  0,
	}

	buffer := make([]byte, chunkSize)
	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading file: %v", err)
		}

		chunk := &pkg.FileChunk{
			ChunkData:  buffer[:n],
			ChunkIndex: firstChunk.ChunkIndex,
			IsLast:     false,
		}

		if firstChunk.ChunkIndex == 0 {
			chunk.FileName = firstChunk.FileName
			chunk.ContentType = firstChunk.ContentType
			chunk.TotalSize = firstChunk.TotalSize
		}

		if err := stream.Send(chunk); err != nil {
			return nil, fmt.Errorf("failed to send chunk: %v", err)
		}

		firstChunk.ChunkIndex++
	}

	// Send last empty chunk to signal completion
	if err := stream.Send(&pkg.FileChunk{
		ChunkIndex: firstChunk.ChunkIndex,
		rr := stream.Send(&pkg.FileChunk{
	    ChunkIndex: firstChunk.ChunkIndex,
		IsLast:     true,
	}); err != nil {
        return nil, fmt.Errorf("failed to send final chunk: %v", err)
	}

    return stream.CloseAndRecv()
}
	
	 (g *GRPCClient) HealthCheck(ctx context.Context) (*pkg.HealthCheckResponse, error) {
ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
    
    return g.client.HealthCheck(ctx, &pkg.HealthCheckRequest{})
}

func main() {
    ctx := context.Background()
    
    // Initialize gRPC client
    grpcClient, err := NewGRPCClient("your-server-address:port")
    if err != nil {
        log.Fatalf("Failed to create gRPC client: %v", err)
    }
    defer grpcClient.Close()

    // Initialize RTSP client
    rtspClient := gortsplib.Client{}
    
    // Parse URL
    url, err := base.ParseURL("rtsp://localhost:8554/stream")
    if err != nil {
        log.Fatalf("Failed to parse URL: %v", err)
    }

    // Connect to RTSP stream
    err = rtspClient.Start(url.Scheme, url.Host)
    if err != nil {
        log.Fatalf("Failed to start RTSP client: %v", err)
    }
    defer rtspClient.Close()

    desc, _, err := rtspClient.Describe(url)
    if err != nil {
        log.Fatalf("Failed to describe RTSP stream: %v", err)
    }

    err = rtspClient.SetupAll(desc.BaseURL, desc.Medias)
    if err != nil {
        log.Fatalf("Failed to setup RTSP medias: %v", err)
    }

    audioHandler := audio.NewAudioHandler()

    // Create a ticker for periodic WAV file creation and upload
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    go func() {
        for range ticker.C {
            outputPath := fmt.Sprintf("audio_%d.wav", time.Now().Unix())
            if err := audioHandler.ConvertToWav(outputPath); err != nil {
                log.Printf("Failed to convert audio: %v", err)
                continue
            }

            // Upload the WAV file
            status, err := grpcClient.UploadWavFile(ctx, outputPath)
            if err != nil {
                log.Printf("Failed to upload WAV file: %v", err)
                continue
            }

            if status.Success {
                log.Printf("Successfully uploaded file to: %s", status.S3Path)
                // Clean up local file after successful upload
                if err := os.Remove(outputPath); err != nil {
                    log.Printf("Failed to remove temporary file: %v", err)
                }
            } else {
                log.Printf("Upload failed: %s", status.Message)
            }
        }
    }()

    rtspClient.OnPacketRTPAny(func(media *description.Media, forma format.Format, pkt *rtp.Packet) {
        if media.Type == "audio" {
            audioHandler.HandlePacket(media, forma, pkt)
        }
    })

    // Start playing the RTSP stream
    if _, err = rtspClient.Play(nil); err != nil {
        log.Fatalf("Failed to play RTSP stream: %v", err)
    }

    // Wait for interrupt signal
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt)
    <-sigChan
}
