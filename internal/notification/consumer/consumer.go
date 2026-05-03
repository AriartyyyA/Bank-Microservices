package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/AriartyyyA/gobank/pkg/kafka"
	"github.com/AriartyyyA/gobank/pkg/kafka/events"
)

type TransferConsumer struct {
	consumer *kafka.Consumer
}

func NewTransferConsumer(consumer *kafka.Consumer) *TransferConsumer {
	return &TransferConsumer{consumer: consumer}
}

func (c *TransferConsumer) Start(ctx context.Context) error {
	for {
		message, err := c.consumer.Read(ctx)
		if err != nil {
			select {
			case <-ctx.Done():
				return fmt.Errorf("ctx close: %w", err)
			default:
				return err
			}
		}

		var event events.TransferEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			log.Printf("failed to unmarshall event: %v", err)
			continue
		}

		log.Printf("transfer event: from=%s to=%s amount=%d",
			event.FromWalletID, event.ToWalletID, event.Amount)
	}
}
