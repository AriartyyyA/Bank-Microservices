package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/AriartyyyA/gobank/pkg/kafka"
	"github.com/AriartyyyA/gobank/pkg/kafka/events"
	"go.opentelemetry.io/otel"
)

type TransferConsumer struct {
	consumer *kafka.Consumer
}

func NewTransferConsumer(consumer *kafka.Consumer) *TransferConsumer {
	return &TransferConsumer{consumer: consumer}
}

func (c *TransferConsumer) Start(ctx context.Context) error {
	for {
		msgCtx, message, err := c.consumer.Read(ctx)
		if err != nil {
			return fmt.Errorf("message read error: %w", err)
		}

		msgCtx, span := otel.Tracer("notification-consumer").Start(msgCtx, "process transfer event")

		var event events.TransferEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			log.Printf("failed to unmarshall event: %v", err)
			span.End()
			continue
		}

		log.Printf("transfer event: from=%s to=%s amount=%d",
			event.FromWalletID, event.ToWalletID, event.Amount)

		span.End()
	}
}
