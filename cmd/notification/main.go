package main

import (
	"context"
	"log"

	"github.com/AriartyyyA/gobank/internal/notification/consumer"
	"github.com/AriartyyyA/gobank/pkg/kafka"
)

func main() {
	c := kafka.NewConsumer([]string{"localhost:9092"}, "transfers", "notification-service")
	transferConsumer := consumer.NewTransferConsumer(c)

	if err := transferConsumer.Start(context.Background()); err != nil {
		log.Println(err)
	}
}
