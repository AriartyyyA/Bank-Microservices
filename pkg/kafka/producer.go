package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	return &Producer{writer: writer}
}

func (p *Producer) Publish(ctx context.Context, key string, value []byte) error {
	message := kafka.Message{
		Key:   []byte(key),
		Value: value,
	}

	if err := p.writer.WriteMessages(ctx, message); err != nil {
		return err
	}

	return nil
}

func (p *Producer) Close() error {
	if err := p.writer.Close(); err != nil {
		return err
	}

	return nil
}
