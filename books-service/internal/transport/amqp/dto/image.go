package dto

import "github.com/google/uuid"

type ImageCompressMessage struct {
	BookID  uuid.UUID `json:"book_id"`
	CoverID uuid.UUID `json:"cover_id"`
	Path    string    `json:"original_path"`
}
