package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	cfg := kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	}

	reader := kafka.NewReader(cfg)

	return &Consumer{reader: reader}
}

func (c *Consumer) Read(ctx context.Context) (kafka.Message, error) {
	message, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return kafka.Message{}, err
	}

	return message, nil
}

func (c *Consumer) Close() error {
	if err := c.reader.Close(); err != nil {
		return err
	}

	return nil
}
