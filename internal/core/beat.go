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
		AddBeat(ctx context.Context, userID int, beatPath string) (beatID int, err error)
		GetPresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error)
	}

	BeatService interface {
		GetBeat(ctx context.Context, beatID int64, start int64, end *int64) (obj *minio.Object, size int64, contentType string, err error)
		AddBeat(ctx context.Context, userID int) (beatID int, beatPath string, err error)
		WritePartialContent(ctx context.Context, r io.Reader, w io.Writer, chinkSize int) error
		GetUploadURL(ctx context.Context, beatPath string) (string, error)
	}

	Beat struct {
		ID        int
		UserID    int
		Path      string
		IsDeleted bool
		CreatedAt time.Time
		UpdatedAt time.Time
	}
)
