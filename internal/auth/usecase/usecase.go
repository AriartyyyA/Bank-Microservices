package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AriartyyyA/gobank/internal/auth/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase struct {
	repo      domain.UserRepository
	jwtSecret string
}

func NewAuthUseCase(repo domain.UserRepository, jwtSecret string) *AuthUseCase {
	return &AuthUseCase{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (u *AuthUseCase) Register(ctx context.Context, email, password string) error {
	_, err := u.repo.GetUserByEmail(ctx, email)
	if err == nil {
		return domain.ErrUserExists
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return fmt.Errorf("check user exists: %w", err)
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	id := uuid.New().String()

	user := domain.User{
		UUID:         id,
		Email:        email,
		PasswordHash: string(hashPassword),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := u.repo.CreateUser(ctx, user); err != nil {
		return fmt.Errorf("register user: %w", err)
	}

	return nil
}

func (u *AuthUseCase) Login(ctx context.Context, email, password string) (string, error) {
	user, err := u.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", domain.ErrUserNotFound
		}

		return "", fmt.Errorf("get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", fmt.Errorf("wrong password: %w", domain.ErrWrongPassword)
	}

	return u.generateJWT(user.UUID, user.Email)
}

func (u *AuthUseCase) ValidateToken(token string) (userID, email string, err error) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}

		return []byte(u.jwtSecret), nil
	})
	if err != nil {
		return "", "", fmt.Errorf("parse token: %w", err)
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return "", "", fmt.Errorf("invalid token")
	}

	userID = claims["userID"].(string)
	email = claims["email"].(string)

	return userID, email, nil
}

func (u *AuthUseCase) generateJWT(userID, email string) (string, error) {
	claims := jwt.MapClaims{
		"userID": userID,
		"email":  email,
		"exp":    time.Now().Add(time.Minute * 15).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(u.jwtSecret))
}
