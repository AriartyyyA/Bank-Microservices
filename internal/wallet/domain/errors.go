package domain

import "errors"

var (
	ErrWalletNotFound            = errors.New("wallet not found")
	ErrNoMoney                   = errors.New("low balance")
	ErrTransactionNegativeAmount = errors.New("negative amount")
	ErrToSendMyself              = errors.New("to send amount myself")
	ErrWalletIsExists            = errors.New("wallet is exists")
)
