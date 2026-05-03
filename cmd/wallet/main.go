package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	transport "github.com/AriartyyyA/gobank/internal/wallet/delivery/http"
	grpcClient "github.com/AriartyyyA/gobank/internal/wallet/grpc"
	pg_repo "github.com/AriartyyyA/gobank/internal/wallet/repository/pg"
	"github.com/AriartyyyA/gobank/internal/wallet/usecase"
	"github.com/AriartyyyA/gobank/pkg/kafka"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

func main() {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DB_URL_WALLET"))
	if err != nil {
		log.Fatal(err)
	}

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")

	repo := pg_repo.NewPostgresRepo(pool)
	producer := kafka.NewProducer(brokers, "transfers")
	uc := usecase.NewWalletUseCase(repo, producer)
	handlers := transport.NewHandlerWallet(uc)

	authClient, err := grpcClient.NewAuthClient(os.Getenv("AUTH_GRPC_ADDR"))
	if err != nil {
		log.Fatal(err)
	}

	router := chi.NewRouter()
	router.Use(transport.GRPCAuthMiddleware(authClient))
	handlers.RegisterRoutes(router)

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
