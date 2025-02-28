package model

import (
	"errors"
	"fmt"
)

var (
	ErrBeatNotFound       = errors.New("beat not found")
	ErrInvalidRangeHeader = errors.New("invalid range header")
	ErrBeatAlreadyExists  = errors.New("beat already exists")
	ErrValidationFailed   = errors.New("validation failed")
	ErrInvalidMediaType   = errors.New("invalid media type: must be one of file, archive or image")
	ErrSizeExceeded       = errors.New("size exceeded")
	ErrInvalidHash        = errors.New("invalid hash")
	ErrInvalidExpiry      = errors.New("invalid expiry")
	ErrURLExpired         = errors.New("url expired")
	ErrArchiveNotFound    = errors.New("archive not found")
	ErrOwnerNotFound      = errors.New("owner not found")
	ErrInvalidOwner       = errors.New("invalid owner")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInvalidID          = errors.New("invalid id")
)

type ModelError struct {
	Err error
	Msg string
}

func (me *ModelError) Error() string {
	if me.Msg == "" {
		return me.Err.Error()
	}
	return fmt.Sprintf("%s: %s", me.Err.Error(), me.Msg)
}

func (me *ModelError) Unwrap() error {
	return me.Err
}

func NewErr(err error, msg string) *ModelError {
	return &ModelError{
		Err: err,
		Msg: msg,
	}
}
