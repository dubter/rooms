package domain

import "errors"

var (
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrNicknameAlreadyExist = errors.New("nickname already exist")
	ErrUserNotFound         = errors.New("user not found by refresh token")
	ErrRoomNotFound         = errors.New("room not found")
)
