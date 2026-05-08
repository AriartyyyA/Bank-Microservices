package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/AriartyyyA/gobank/internal/notification/consumer"
	"github.com/AriartyyyA/gobank/pkg/kafka"
	"github.com/AriartyyyA/gobank/pkg/tracing"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM, syscall.SIGINT,
	)
	defer cancel()

	tracingShutdown, err := tracing.Init(ctx, "notification-service")
	if err != nil {
		log.Fatalf("tracing error: %v", err)
	}

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	c := kafka.NewConsumer(brokers, "transfers", "notification-service")
	transferConsumer := consumer.NewTransferConsumer(c)

	go func() {
		if err := transferConsumer.Start(ctx); err != nil {
			log.Println(err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := tracingShutdown(shutdownCtx); err != nil {
		log.Printf("tracing shutdown error: %v", err)
	}

	log.Println("shutting down")
	c.Close()
}
