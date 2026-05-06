package domain

import "errors"

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrUserExists          = errors.New("user already exists")
	ErrWrongPassword       = errors.New("wrong password")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)
