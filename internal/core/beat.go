package core

import (
	"context"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
)

type (
	BeatStorage interface {
		GetBeatFromS3(ctx context.Context, beatPath string, start, end int64) (*minio.Object, int, error)
		GetBeatByID(ctx context.Context, id int) (*Beat, error)
	}

	BeatService interface {
		GetBeat(ctx context.Context, beatID int, start, end int64) (obj *minio.Object, size int, err error)
		WritePartialContent(ctx context.Context, r io.Reader, w io.Writer, chinkSize int) error
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
