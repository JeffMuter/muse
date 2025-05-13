package server

import (
	"context"
	"time"

	pb "github.com/jeffmuter/muse/proto"
)

func (s *OwlServer) SaveTranscript(ctx context.Context, req *pb.SaveTranscriptRequest) (*pb.SaveTranscriptResponse, error) {
	err := s.dbService.SaveTranscript(ctx, req.TranscriptId, req.TranscriptText, req.Summary)
	if err != nil {
		return &pb.SaveTranscriptResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.SaveTranscriptResponse{
		Success: true,
		Message: "Transcript saved successfully",
	}, nil
}

func (s *OwlServer) GetTranscript(ctx context.Context, req *pb.GetTranscriptRequest) (*pb.GetTranscriptResponse, error) {
	transcript, err := s.dbService.GetTranscript(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.GetTranscriptResponse{
		Id:             transcript.ID,
		TranscriptId:   transcript.TranscriptID,
		TranscriptText: transcript.TranscriptText,
		Summary:        transcript.Summary,
		CreatedAt:      transcript.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:      transcript.UpdatedAt.Time.Format(time.RFC3339),
	}, nil
}
