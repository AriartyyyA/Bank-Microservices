package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/AriartyyyA/gobank/internal/notification/consumer"
	"github.com/AriartyyyA/gobank/pkg/kafka"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM, syscall.SIGINT,
	)
	defer cancel()

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	c := kafka.NewConsumer(brokers, "transfers", "notification-service")
	transferConsumer := consumer.NewTransferConsumer(c)

	go func() {
		if err := transferConsumer.Start(ctx); err != nil {
			log.Println(err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down")
	c.Close()
}
