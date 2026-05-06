package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/AriartyyyA/gobank/internal/auth/domain"
	"github.com/AriartyyyA/gobank/internal/auth/domain/mocks"
	"github.com/AriartyyyA/gobank/internal/auth/usecase"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func setupRedis(t *testing.T) *redis.Client {
	mr := miniredis.RunT(t)
	return redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
}

func TestAuthUseCase_Register_Success(t *testing.T) {
	mockRepo := mocks.NewUserRepository(t)

	mockRepo.On("GetUserByEmail", mock.Anything, "test@test.com").
		Return(nil, domain.ErrUserNotFound)

	mockRepo.On("CreateUser", mock.Anything, mock.Anything).Return(nil)

	uc := usecase.NewAuthUseCase(mockRepo, "secret", setupRedis(t))
	err := uc.Register(context.Background(), "test@test.com", "password123")

	assert.NoError(t, err)
}

func TestAuthUseCase_Register_AlreadyExists(t *testing.T) {
	mockRepo := mocks.NewUserRepository(t)

	mockRepo.On("GetUserByEmail", mock.Anything, "test@test.com").Return(&domain.User{}, nil)

	uc := usecase.NewAuthUseCase(mockRepo, "secret", setupRedis(t))
	err := uc.Register(context.Background(), "test@test.com", "password123")

	assert.ErrorIs(t, err, domain.ErrUserExists)
}

func TestAuthUseCase_Register_DBError(t *testing.T) {
	mockRepo := mocks.NewUserRepository(t)

	mockRepo.On("GetUserByEmail", mock.Anything, "test@test.com").Return(nil, errors.New("db error"))

	uc := usecase.NewAuthUseCase(mockRepo, "secret", setupRedis(t))
	err := uc.Register(context.Background(), "test@test.com", "password123")

	assert.Error(t, err)
}

func TestAuthUseCase_Login_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &domain.User{
		UUID:         "some-uuid",
		Email:        "test@test.com",
		PasswordHash: string(hash),
	}

	mockRepo := mocks.NewUserRepository(t)

	mockRepo.On("GetUserByEmail", mock.Anything, "test@test.com").Return(user, nil)

	uc := usecase.NewAuthUseCase(mockRepo, "secret", setupRedis(t))
	accessToken, refreshToken, err := uc.Login(context.Background(), "test@test.com", "password123")

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
}

func TestAuthUserCase_Login_UserNotFound(t *testing.T) {
	mockRepo := mocks.NewUserRepository(t)

	mockRepo.On("GetUserByEmail", mock.Anything, "test@test.com").Return(nil, domain.ErrUserNotFound)

	uc := usecase.NewAuthUseCase(mockRepo, "secret", setupRedis(t))
	_, _, err := uc.Login(context.Background(), "test@test.com", "password123")

	assert.ErrorIs(t, err, domain.ErrUserNotFound)
}

func TestAuthUserCase_Login_WrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &domain.User{
		UUID:         "some-uuid",
		Email:        "test@test.com",
		PasswordHash: string(hash),
	}

	mockRepo := mocks.NewUserRepository(t)
	mockRepo.On("GetUserByEmail", mock.Anything, "test@test.com").Return(user, nil)

	uc := usecase.NewAuthUseCase(mockRepo, "secret", setupRedis(t))
	_, _, err := uc.Login(context.Background(), "test@test.com", "wrongpassword")

	assert.ErrorIs(t, err, domain.ErrWrongPassword)
}
