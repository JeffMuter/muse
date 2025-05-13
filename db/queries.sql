-- name: GetTranscriptById :one
SELECT * FROM transcripts WHERE id = ?;
