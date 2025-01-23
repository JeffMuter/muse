package main

import (
	"context"
	"fmt"
	"io"
	"log"
	pb "muse/pelican/pkg/file"
	"net"
	"os"
	"path/filepath"

	"google.golang.org/grpc"
)

type testServer struct {
	pb.UnimplementedAudioFileServiceServer
	outputDir string
}

func (s *testServer) UploadFile(stream pb.AudioFileService_UploadFileServer) error {
	var (
		fileData    []byte
		fileName    string
		contentType string
	)

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to receive chunk: %v", err)
		}

		// Store metadata from first chunk
		if len(fileData) == 0 {
			fileName = chunk.FileName
			contentType = chunk.ContentType
			log.Printf("Starting to receive file: %s (type: %s, size: %d bytes)",
				fileName, contentType, chunk.TotalSize)
		}

		fileData = append(fileData, chunk.ChunkData...)
		log.Printf("Received chunk %d, size: %d bytes", chunk.ChunkIndex, len(chunk.ChunkData))
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(s.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Save the complete file
	outputPath := filepath.Join(s.outputDir, fileName)
	if err := os.WriteFile(outputPath, fileData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	log.Printf("Successfully saved file to: %s", outputPath)

	// Send success response
	return stream.SendAndClose(&pb.UploadStatus{
		Success:      true,
		Message:      "File uploaded successfully",
		FileId:       fileName,
		ReceivedSize: int64(len(fileData)),
		S3Path:       outputPath,
	})
}

func (s *testServer) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Status: pb.HealthCheckResponse_SERVING,
	}, nil
}

func main() {
	port := 50051
	outputDir := "received_files"

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterAudioFileServiceServer(server, &testServer{outputDir: outputDir})

	log.Printf("Starting test server on port %d", port)
	log.Printf("Files will be saved to: %s", outputDir)

	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
