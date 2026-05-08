package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
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

func (c *Consumer) Read(ctx context.Context) (context.Context, kafka.Message, error) {
	message, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return nil, kafka.Message{}, err
	}

	carrier := propagation.MapCarrier{}
	for _, h := range message.Headers {
		carrier[h.Key] = string(h.Value)
	}
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

	return ctx, message, nil
}

func (c *Consumer) Close() error {
	if err := c.reader.Close(); err != nil {
		return err
	}

	return nil
}
