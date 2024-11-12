package core

import (
	"context"
	"io"
	"time"
)

type (
	BeatStorage interface {
		GetBeatFromS3(ctx context.Context, beatPath string, start int64, end *int64) (obj io.ReadCloser, size int64, contentType string, err error)
		GetBeatByID(ctx context.Context, id int64, param IsDownloaded) (*Beat, error)
		AddBeat(ctx context.Context, beat Beat, beatGenre []BeatGenre) (beatID int, err error)
		GetPresignedURL(ctx context.Context, beatPath string, expiry time.Duration) (url string, err error)
		GetBeatByFilter(ctx context.Context, filter FeedFilter, seen []string) (beat *Beat, genre *string, err error)
		GetBeatGenres(ctx context.Context, beatID int) (beatGenres []BeatGenre, err error)
		GetBeatsByBeatmakerID(ctx context.Context, beatmakerID int, p GetBeatsParams) (beats []Beat, total int, err error)

		GetUserSeenBeats(ctx context.Context, userID int) ([]string, error)
		AddUserSeenBeat(ctx context.Context, userID int, beatID int) error
		PopUserSeenBeat(ctx context.Context, userID int) error
		ClearUserSeenBeats(ctx context.Context, userID int) error
	}

	BeatService interface {
		GetBeatFromS3(ctx context.Context, beatID int64, start int64, end *int64) (obj io.ReadCloser, size int64, contentType string, err error)
		AddBeat(ctx context.Context, beat Beat, beatGenre []BeatGenre) (beatPath string, err error)
		WritePartialContent(ctx context.Context, r io.Reader, w io.Writer, chunkSize int) error
		GetUploadURL(ctx context.Context, beatPath string) (url string, err error)
		GetBeatByFilter(ctx context.Context, userID int, params FeedFilter) (beat *Beat, genre *string, err error)
		GetBeat(ctx context.Context, beatID int) (beat *Beat, beatGenres []BeatGenre, err error)
		GetBeatsByBeatmakerID(ctx context.Context, beatmakerID int, p GetBeatsParams) (beats []Beat, beatsGenres [][]BeatGenre, total int, err error)
	}

	FeedFilter struct {
		Genre string
	}

	BeatGenre struct {
		ID     int
		BeatID int
		Genre  string
	}

	Beat struct {
		ID           int
		BeatmakerID  int
		Path         string
		Name         string
		Description  string
		IsDownloaded bool
		IsDeleted    bool
		CreatedAt    time.Time
		UpdatedAt    time.Time
	}

	GetBeatsParams struct {
		Limit  int
		Offset int
		Order  string
	}

	IsDownloaded int
)

const (
	True IsDownloaded = iota
	False
	Any
)
