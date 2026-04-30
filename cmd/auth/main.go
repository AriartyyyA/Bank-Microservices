package main

import (
	"context"
	"log"
	"net/http"
	"os"

	transport "github.com/AriartyyyA/gobank/internal/auth/delivery/http"
	pg_repo "github.com/AriartyyyA/gobank/internal/auth/repository/pg"
	"github.com/AriartyyyA/gobank/internal/auth/usecase"
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
	pool, err := pgxpool.New(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal(err)
	}

	repo := pg_repo.NewPostgresRepo(pool)
	uc := usecase.NewAuthUseCase(repo, os.Getenv("JWT_SECRET"))
	handlers := transport.NewHandlerAuth(uc)

	router := chi.NewRouter()
	handlers.RegisterRoutes(router)

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
