package usecase_test

import (
	"context"
	"testing"

	"github.com/AriartyyyA/gobank/internal/wallet/domain"
	"github.com/AriartyyyA/gobank/internal/wallet/domain/mocks"
	"github.com/AriartyyyA/gobank/internal/wallet/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWalletUseCase_Transfer_Success(t *testing.T) {
	mockRepo := mocks.NewWalletRepository(t)

	fromWallet := &domain.Wallet{
		ID:      "from",
		Balance: 500,
	}
	toWallet := &domain.Wallet{
		ID:      "to",
		Balance: 0,
	}

	mockRepo.On("FindWalletByID", mock.Anything, "from").
		Return(fromWallet, nil)

	mockRepo.On("FindWalletByID", mock.Anything, "to").
		Return(toWallet, nil)

	mockRepo.On("WithTx", mock.Anything, mock.Anything).
		Return(nil).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			fn(context.Background())
		})

	mockRepo.On("UpdateBalance", mock.Anything, "from", int64(-100)).
		Return(nil)
	mockRepo.On("UpdateBalance", mock.Anything, "to", int64(100)).
		Return(nil)

	mockRepo.On("CreateTransaction", mock.Anything, mock.Anything).
		Return(nil)

	uc := usecase.NewWalletUseCase(mockRepo)
	err := uc.Transfer(
		context.Background(),
		fromWallet.ID,
		toWallet.ID,
		100,
	)

	assert.NoError(t, err)
}

func TestWalletUseCase_Transfer_NegativeAmount(t *testing.T) {
	fromWallet := &domain.Wallet{
		ID:      "from",
		Balance: 500,
	}
	toWallet := &domain.Wallet{
		ID:      "to",
		Balance: 0,
	}

	mockRepo := mocks.NewWalletRepository(t)
	uc := usecase.NewWalletUseCase(mockRepo)
	err := uc.Transfer(
		context.Background(),
		fromWallet.ID,
		toWallet.ID,
		-100,
	)

	assert.ErrorIs(t, err, domain.ErrTransactionNegativeAmount)
}

func TestWalletUseCase_Transfer_SameWallet(t *testing.T) {
	fromWallet := &domain.Wallet{
		ID:      "from",
		Balance: 500,
	}
	toWallet := &domain.Wallet{
		ID:      "from",
		Balance: 0,
	}

	mockRepo := mocks.NewWalletRepository(t)
	uc := usecase.NewWalletUseCase(mockRepo)
	err := uc.Transfer(
		context.Background(),
		fromWallet.ID,
		toWallet.ID,
		100,
	)

	assert.ErrorIs(t, err, domain.ErrToSendMyself)
}

func TestWalletUseCase_Transfer_NoMoney(t *testing.T) {
	fromWallet := &domain.Wallet{
		ID:      "from",
		Balance: 200,
	}
	toWallet := &domain.Wallet{
		ID:      "to",
		Balance: 0,
	}

	mockRepo := mocks.NewWalletRepository(t)
	mockRepo.On("FindWalletByID", mock.Anything, "from").
		Return(fromWallet, nil)

	uc := usecase.NewWalletUseCase(mockRepo)
	err := uc.Transfer(
		context.Background(),
		fromWallet.ID,
		toWallet.ID,
		500,
	)

	assert.ErrorIs(t, err, domain.ErrNoMoney)
}

func TestWalletUseCase_Transfer_WalletNotFound(t *testing.T) {
	mockRepo := mocks.NewWalletRepository(t)
	mockRepo.On("FindWalletByID", mock.Anything, "from").
		Return(nil, domain.ErrWalletNotFound)

	uc := usecase.NewWalletUseCase(mockRepo)
	err := uc.Transfer(
		context.Background(),
		"from",
		"to",
		500,
	)

	assert.ErrorIs(t, err, domain.ErrWalletNotFound)
}
