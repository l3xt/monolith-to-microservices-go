package amqp

import (
	"bookshelf/books-service/internal/transport/amqp/dto"
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

const ( 
	QueueImageCompress = "image_compress"
)

type Client interface {
	PublishMessage(ctx context.Context, queueName string, message []byte) error
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
		return err
	}
	
	return p.client.PublishMessage(ctx, QueueImageCompress, message)
}