package model

import (
	"errors"
	"fmt"
)

var (
	ErrBeatNotFound        = errors.New("beat not found")
	ErrOrderByInvalidField = errors.New("order_by: invalid field")
	ErrInvalidGenreID      = errors.New("invalid genre id")
	ErrInvalidTagID        = errors.New("invalid tag id")
	ErrInvalidMoodID       = errors.New("invalid mood id")
	ErrInvalidNoteID       = errors.New("invalid note id")
	ErrInvalidRangeHeader  = errors.New("invalid range header")
	ErrBeatAlreadyExists   = errors.New("beat already exists")
	ErrValidationFailed    = errors.New("validation failed")
	ErrInvalidBeatID       = errors.New("invalid beat id")
	ErrInvalidType         = errors.New("invalid type: must be one of file, archive or image")
	ErrFileSizeExceeded    = errors.New("file size exceeded")
	ErrArchiveSizeExceeded = errors.New("archive size exceeded")
	ErrImageSizeExceeded   = errors.New("image size exceeded")
	ErrInvalidHash         = errors.New("invalid hash")
	ErrInvalidExpiry       = errors.New("invalid expiry")
	ErrURLExpired          = errors.New("url expired")
	ErrArchiveNotFound     = errors.New("archive not found")
	ErrOwnerNotFound       = errors.New("owner not found")
	ErrInvalidOwner        = errors.New("invalid owner")
	ErrUnauthorized        = errors.New("unauthorized")
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
