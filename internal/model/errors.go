package model

import "errors"

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
)
