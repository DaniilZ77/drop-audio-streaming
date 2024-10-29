package core

import "errors"

var (
	ErrBeatNotFound = errors.New("beat not found")
	ErrInvalidRange = errors.New("invalid range")
	ErrInternal     = errors.New("internal error")
	ErrUnauthorized = errors.New("unauthorized")
)
