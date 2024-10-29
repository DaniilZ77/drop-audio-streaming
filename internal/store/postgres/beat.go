package beat

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/minio"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/postgres"
	miniolib "github.com/minio/minio-go/v7"
)

type store struct {
	*minio.Minio
	*postgres.Postgres
	bucketName string
}

func New(m *minio.Minio, pg *postgres.Postgres, bucketName string) core.BeatStorage {
	return &store{m, pg, bucketName}
}

func (s *store) GetBeatByID(ctx context.Context, id int64) (*core.Beat, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var beat core.Beat

	stmt := `SELECT id, user_id, path, is_deleted, created_at, updated_at FROM beats WHERE id = $1`

	err := s.DB.QueryRowContext(ctx, stmt, id).Scan(&beat.ID, &beat.UserID, &beat.Path, &beat.IsDeleted, &beat.CreatedAt, &beat.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, core.ErrBeatNotFound
		}
	}

	return &beat, nil
}

func (s *store) GetBeatFromS3(ctx context.Context, beatPath string, start int64, end *int64) (*miniolib.Object, int64, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	objInfo, err := s.Client.StatObject(ctx, s.bucketName, beatPath, miniolib.StatObjectOptions{})
	if err != nil {
		return nil, 0, "", err
	}

	if start >= objInfo.Size {
		return nil, 0, "", core.ErrInvalidRange
	}

	if *end >= objInfo.Size {
		*end = objInfo.Size - 1
	}

	opts := miniolib.GetObjectOptions{}
	if start != 0 || *end != -1 {
		opts.SetRange(start, *end)
	}

	obj, err := s.Client.GetObject(ctx, s.bucketName, beatPath, opts)
	if err != nil {
		return nil, 0, "", err
	}

	return obj, objInfo.Size, objInfo.ContentType, nil
}

func (s *store) AddBeat(ctx context.Context, userID int, beatPath string) (beatID int, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt := `INSERT INTO beats (user_id, path) VALUES ($1, $2) RETURNING id`
	err = s.DB.QueryRowContext(ctx, stmt, userID, beatPath).Scan(&beatID)
	if err != nil {
		return 0, err
	}

	return beatID, nil
}

func (s *store) GetPresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	u, err := s.Client.PresignedPutObject(ctx, s.bucketName, objectName, expiry)
	if err != nil {
		return "", err
	}

	return u.String(), nil
}
