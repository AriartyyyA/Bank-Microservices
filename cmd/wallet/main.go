package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/AriartyyyA/gobank/docs/wallet"
	transport "github.com/AriartyyyA/gobank/internal/wallet/delivery/http"
	grpcClient "github.com/AriartyyyA/gobank/internal/wallet/grpc"
	pg_repo "github.com/AriartyyyA/gobank/internal/wallet/repository/pg"
	"github.com/AriartyyyA/gobank/internal/wallet/usecase"
	"github.com/AriartyyyA/gobank/pkg/kafka"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

// @title Bank Wallet API
// @version 1.0
// @description Сервис переводов и кошельков
// @host localhost:8081
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM, syscall.SIGINT,
	)
	defer cancel()

	pool, err := pgxpool.New(context.Background(), os.Getenv("DB_URL_WALLET"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")

	repo := pg_repo.NewPostgresRepo(pool)

	producer := kafka.NewProducer(brokers, "transfers")
	defer producer.Close()

	uc := usecase.NewWalletUseCase(repo, producer)
	handlers := transport.NewHandlerWallet(uc)

	authClient, err := grpcClient.NewAuthClient(os.Getenv("AUTH_GRPC_ADDR"))
	if err != nil {
		log.Fatal(err)
	}

	router := chi.NewRouter()
	router.Use(transport.GRPCAuthMiddleware(authClient))
	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8081/swagger/doc.json"),
	))
	handlers.RegisterRoutes(router)

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		log.Println("HTTP server started on :8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}

	log.Println("server stopped gracefully")
}
