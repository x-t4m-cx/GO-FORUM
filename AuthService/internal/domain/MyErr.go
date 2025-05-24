package domain

import "errors"

var (
	UserNotFound      = errors.New("user not found")
	UserAlreadyExists = errors.New("user already exists")
	InvalidToken      = errors.New("invalid token")
	InvalidData       = errors.New("invalid data")
)
