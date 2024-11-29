package core

import "errors"

var (
	ErrBeatNotFound            = errors.New("beat not found")
	ErrInvalidRange            = errors.New("invalid range")
	ErrInternal                = errors.New("internal error")
	ErrUnauthorized            = errors.New("unauthorized")
	ErrBeatExists              = errors.New("beat already exists")
	ErrValidationFailed        = errors.New("validation failed")
	ErrUnavailable             = errors.New("unavailable")
	ErrInvalidID               = errors.New("invalid id")
	ErrAmountOfRetriesExceeded = errors.New("amount of retries exceeded")
)
