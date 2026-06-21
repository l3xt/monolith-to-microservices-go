package amqp

import (
	"bookshelf/books-service/internal/transport/amqp/dto"
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

const (
	QueueImageCompress = "image_compress"
)

type Client interface {
	PublishMessage(ctx context.Context, queueName string, message []byte) error
	HealthCheck(ctx context.Context) error
	Close() error
}

type Producer struct {
	client Client
}

func NewProducer(client Client) *Producer {
	return &Producer{client: client}
}

func (p *Producer) PublishImageCompressEvent(ctx context.Context, bookID, coverID uuid.UUID, path string) error {
	msg := dto.ImageCompressMessage{
		BookID:  bookID,
		CoverID: coverID,
		Path:    path,
	}
	message, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("Producer.PublishImageCompressEvent: %w", err)
	}

	if err := p.client.PublishMessage(ctx, QueueImageCompress, message); err != nil {
		return fmt.Errorf("Producer.PublishImageCompressEvent: %w", err)
	}

	return nil
}

func (p *Producer) HealthCheck(ctx context.Context) error {
	if err := p.client.HealthCheck(ctx); err != nil {
		return fmt.Errorf("Producer.HealthCheck: %w", err)
	}
	return nil
}

func (p *Producer) Close() error {
	if err := p.client.Close(); err != nil {
		return fmt.Errorf("Producer.Close: %w", err)
	}
	return nil
}
