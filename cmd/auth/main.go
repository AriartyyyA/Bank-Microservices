package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/AriartyyyA/gobank/docs/auth"
	grpcDelivery "github.com/AriartyyyA/gobank/internal/auth/delivery/grpc"
	transport "github.com/AriartyyyA/gobank/internal/auth/delivery/http"
	pg_repo "github.com/AriartyyyA/gobank/internal/auth/repository/pg"
	"github.com/AriartyyyA/gobank/internal/auth/usecase"
	"github.com/AriartyyyA/gobank/pkg/ratelimit"
	pb "github.com/AriartyyyA/gobank/proto/auth"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/grpc"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

// @title Bank Auth API
// @version 1.0
// @description Сервис авторизации
// @host localhost:8080
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

	pool, err := pgxpool.New(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	jwtSecret := os.Getenv("JWT_SECRET")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: os.Getenv("REDIS_PASSWORD"),
	})
	defer redisClient.Close()

	rate := ratelimit.NewRateLimit(redisClient, 100, time.Minute)

	repo := pg_repo.NewPostgresRepo(pool)
	uc := usecase.NewAuthUseCase(repo, jwtSecret, redisClient)

	// grpc server
	grpcServer := grpc.NewServer()
	authGrpc := grpcDelivery.NewAuthGRPCServer(uc)
	pb.RegisterAuthServiceServer(grpcServer, authGrpc)

	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		log.Println("gRPC server started on :9090")
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	// server
	handlers := transport.NewHandlerAuth(uc, jwtSecret)

	router := chi.NewRouter()
	router.Use(ratelimit.Middleware(rate))
	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
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

	grpcServer.GracefulStop()

	log.Println("servers stopped gracefully")
}
