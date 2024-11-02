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
		AddBeat(ctx context.Context, beat Beat) (beatID int, err error)
		GetPresignedURL(ctx context.Context, path string, expiry time.Duration) (url string, err error)
		GetBeatByParams(ctx context.Context, params BeatParams, seen []string) (beat *Beat, err error)
		GetUserSeenBeats(ctx context.Context, userID int) ([]string, error)
		AddUserSeenBeat(ctx context.Context, userID int, beatID int) error
		PopUserSeenBeat(ctx context.Context, userID int) error
		ClearUserSeenBeats(ctx context.Context, userID int) error
	}

	BeatService interface {
		GetBeat(ctx context.Context, beatID int64, start int64, end *int64) (obj io.ReadCloser, size int64, contentType string, err error)
		AddBeat(ctx context.Context, beat Beat) (beatPath string, err error)
		WritePartialContent(ctx context.Context, r io.Reader, w io.Writer, chinkSize int) error
		GetUploadURL(ctx context.Context, beatPath string) (url string, err error)
		GetBeatByParams(ctx context.Context, userID int, params BeatParams) (beat *Beat, err error)
	}

	BeatParams struct {
		Artist string
		Genre  string
	}

	Beat struct {
		ExternalID  int
		BeatmakerID int
		Path        string
		Artist      string
		Genre       string
		IsDeleted   bool
		CreatedAt   time.Time
		UpdatedAt   time.Time
	}
)
