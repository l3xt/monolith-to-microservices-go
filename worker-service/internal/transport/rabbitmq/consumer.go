package rabbitmq

import (
	"bookshelf/pkg/rabbitmq"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"bookshelf/worker-service/internal/apperror"
	applogger "bookshelf/worker-service/internal/logger"

	amqp "github.com/rabbitmq/amqp091-go"
)

type HandlerFunc func(ctx context.Context, body []byte) error

type Consumer struct {
	client        *rabbitmq.RabbitMQClient
	handlers      map[string]HandlerFunc
	prefetchCount int
	wg            sync.WaitGroup
}

func NewConsumer(client *rabbitmq.RabbitMQClient) (*Consumer, error) {
	return &Consumer{
		client:   client,
		handlers: make(map[string]HandlerFunc),
	}, nil
}

func (c *Consumer) RegisterHandler(queueName string, handler HandlerFunc) {
	c.handlers[queueName] = handler
}

func (c *Consumer) Start(ctx context.Context) error {
	for queue, handler := range c.handlers {
		if err := c.client.DeclareQueue(queue); err != nil {
			return fmt.Errorf("rabbitmq.Consumer.Start: declare queue: %w", err)
		}

		if err := c.client.SetQoS(c.prefetchCount, 0, false); err != nil {
			return fmt.Errorf("rabbitmq.Consumer.Start: set qos: %w", err)
		}

		msgs, err := c.client.Consume(queue)
		if err != nil {
			return fmt.Errorf("rabbitmq.Consumer.Start: consume queue: %w", err)
		}

		c.wg.Add(1)
		go c.consume(ctx, queue, handler, msgs)
	}

	return nil
}

func (c *Consumer) consume(ctx context.Context, queue string, handler HandlerFunc, msgs <-chan amqp.Delivery) {
	defer c.wg.Done()
	log := applogger.FromContext(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Info("context cancelled, stopping consumer", slog.String("queue", queue))
			return
		case msg, ok := <-msgs:
			if !ok {
				log.Warn("rabbitMQ channel is closed", slog.String("queue", queue))
				return
			}

			err := handler(ctx, msg.Body)
			if err == nil {
				log.Info("message processed successfully", slog.String("queue", queue))
				msg.Ack(false)
			} else {
				if errors.Is(err, context.Canceled) {
					log.Warn("processing interrupted by shutdown, requeuing message", slog.String("queue", queue))
					msg.Nack(false, true)	// Возврат
					return
				}

				var retryable *apperror.RetryableError
				if errors.As(err, &retryable) {
					// Ошибка временная, возвращаем в очередь
					log.Warn("temporary error while processing message", slog.Any("error", err))
					msg.Nack(false, true)
				} else {
					// Ошибка постоянная, удаляем (DLQ в будущем)
					log.Warn("permanent error while processing message", slog.Any("error", err))
					msg.Nack(false, false)
				}
			}
		}
	}
}

func (c *Consumer) Wait() {
	c.wg.Wait()
}
