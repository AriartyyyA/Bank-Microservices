package main

import (
	"context"
	"log"
	"net/http"
	"os"

	_ "github.com/AriartyyyA/gobank/docs/auth"
	transport "github.com/AriartyyyA/gobank/internal/auth/delivery/http"
	pg_repo "github.com/AriartyyyA/gobank/internal/auth/repository/pg"
	"github.com/AriartyyyA/gobank/internal/auth/usecase"
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

// @title Bank Auth API
// @version 1.0
// @description Сервис авторизации
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal(err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")

	repo := pg_repo.NewPostgresRepo(pool)
	uc := usecase.NewAuthUseCase(repo, jwtSecret)
	handlers := transport.NewHandlerAuth(uc, jwtSecret)

	router := chi.NewRouter()
	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))
	handlers.RegisterRoutes(router)

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
