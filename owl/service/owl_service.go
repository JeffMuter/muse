package service

import (
	"context"
	"database/sql"
	db "muse/db/generated"
)

type OwlService struct {
	queries *db.Queries
	db      *sql.DB
}

func NewOwlService(database *sql.DB) *OwlService {
	return &OwlService{
		queries: db.New(database),
		db:      database,
	}
}

func (s *OwlService) SaveTranscript(ctx context.Context, transcriptID, text, summary string) error {
	// Use the generated sqlc methods
	_, err := s.queries.CreateTranscript(ctx, db.CreateTranscriptParams{
		TranscriptID:   transcriptID,
		TranscriptText: text,
		Summary:        summary,
	})
	return err
}

func (s *OwlService) GetTranscript(ctx context.Context, id int64) (db.Transcript, error) {
	return s.queries.GetTranscriptById(ctx, id)
}
