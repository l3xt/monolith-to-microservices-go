package handler

import (
	"bookshelf/worker-service/internal/transport/rabbitmq/dto"
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type ImageUseCase interface {
	ProcessBookCover(ctx context.Context, bookID, coverID uuid.UUID, imagePath string) error
}

type ImageHandler struct {
	imageService ImageUseCase
}

func NewImageHandler(imageService ImageUseCase) *ImageHandler {
	return &ImageHandler{
		imageService: imageService,
	}
}

func (h *ImageHandler) HandleImageCompress(ctx context.Context, body []byte) error {
	// Парсим JSON в ImageCompressMessage
	var msg dto.ImageCompressMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return fmt.Errorf("ImageHandler.HandleImageCompress: unmarshal: %w", err)
	}

	if err := h.imageService.ProcessBookCover(ctx, msg.BookID, msg.CoverID, msg.OriginalPath); err != nil {
		return fmt.Errorf("ImageHandler.HandleImageCompress: process book cover: %w", err)
	}

	return nil
}
