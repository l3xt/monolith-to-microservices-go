package rabbitmq

import (
	"context"
	"errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewRabbitMQClient(url string) (*RabbitMQClient, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &RabbitMQClient{
		conn: conn,
		ch:   ch,
	}, nil
}

// Объявить очередь
func (c *RabbitMQClient) DeclareQueue(name string) error {
	_, err := c.ch.QueueDeclare(
		name,
		true,
		false,
		false,
		false,
		nil,
	)
	return err
}

// Закрыть соединение
func (c *RabbitMQClient) Close() error {
	if err := c.CloseChannel(); err != nil {
		return err
	}

	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}

func (c *RabbitMQClient) CloseChannel() error {
	if c.ch != nil {
		return c.ch.Close()
	}
	return nil
}

// Статус
func (c *RabbitMQClient) HealthCheck(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("RabbitMQClient.HealthCheck: context canceled: %w", err)
	}

	if c.conn == nil || c.conn.IsClosed() {
		return errors.New("RabbitMQClient.HealthCheck: connection is closed")
	}

	if c.ch == nil || c.ch.IsClosed() {
		return errors.New("RabbitMQClient.HealthCheck: channel is closed")
	}

	return nil
}


// Публикация сообщения в очередь
func (c *RabbitMQClient) PublishMessage(ctx context.Context, queueName string, message []byte) error {
	err := c.ch.PublishWithContext(
		ctx,
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         message,
		})

	if err != nil {
		return fmt.Errorf("RabbitMQClient.PublishMessage: failed to publish message: %w", err)
	}
	return nil
}

func (c *RabbitMQClient) Consume(queueName string) (<-chan amqp.Delivery, error) {
	deliveries, err := c.ch.Consume(
		queueName,
		"",    // consumer name
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)

	if err != nil {
		return nil, fmt.Errorf("RabbitMQClient.Consume: failed to start consuming: %w", err)
	}
	return deliveries, nil
}

// настраивает правила предварительной выборки сообщений
func (c *RabbitMQClient) SetQoS(prefetchCount int, prefetchSize int, global bool) error {
	err := c.ch.Qos(
		prefetchCount,
		prefetchSize,
		global,
	)

	if err != nil {
		return fmt.Errorf("RabbitMQClient.SetQoS: failed to set QoS: %w", err)
	}
	return nil
}
