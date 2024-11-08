package core

import (
	"context"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
)

type (
	BeatStorage interface {
		GetBeatFromS3(ctx context.Context, beatPath string, start int64, end *int64) (obj *minio.Object, size int64, contentType string, err error)
		GetBeatByID(ctx context.Context, id int64) (*Beat, error)
		AddBeat(ctx context.Context, beat Beat, beatGenre []BeatGenre) (beatID int, err error)
		GetPresignedURL(ctx context.Context, path string, expiry time.Duration) (url string, err error)
		GetBeatByParams(ctx context.Context, params BeatParams, seen []string) (beat *Beat, genre *string, err error)
		GetUserSeenBeats(ctx context.Context, userID int) ([]string, error)
		AddUserSeenBeat(ctx context.Context, userID int, beatID int) error
		PopUserSeenBeat(ctx context.Context, userID int) error
		ClearUserSeenBeats(ctx context.Context, userID int) error
	}

	BeatService interface {
		GetBeat(ctx context.Context, beatID int64, start int64, end *int64) (obj io.ReadCloser, size int64, contentType string, err error)
		AddBeat(ctx context.Context, beat Beat, beatGenre []BeatGenre) (beatPath string, err error)
		WritePartialContent(ctx context.Context, r io.Reader, w io.Writer, chunkSize int) error
		GetUploadURL(ctx context.Context, beatPath string) (url string, err error)
		GetBeatByParams(ctx context.Context, userID int, params BeatParams) (beat *Beat, genre *string, err error)
	}

	BeatParams struct {
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
		IsDownloaded bool
		IsDeleted    bool
		CreatedAt    time.Time
		UpdatedAt    time.Time
	}
)
