package core

import "errors"

var (
	ErrBeatNotFound     = errors.New("beat not found")
	ErrInvalidRange     = errors.New("invalid range")
	ErrInternal         = errors.New("internal error")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrBeatExists       = errors.New("beat already exists")
	ErrValidationFailed = errors.New("validation failed")
	ErrInvalidParams    = errors.New("invalid params")
	ErrUnavailable      = errors.New("unavailable")
	ErrInvalidLimit     = errors.New("invalid limit")
	ErrInvalidOffset    = errors.New("invalid offset")
	ErrInvalidParam     = errors.New("invalid param")
	ErrInvalidID        = errors.New("invalid id")
)
