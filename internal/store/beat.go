package beat

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/minio"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/postgres"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/redis"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	miniolib "github.com/minio/minio-go/v7"
)

type store struct {
	*minio.Minio
	*postgres.Postgres
	bucketName string
	*redis.Redis
	userHistory int
}

func New(
	m *minio.Minio,
	pg *postgres.Postgres,
	bucketName string,
	rdb *redis.Redis,
	userHistory int) core.BeatStorage {
	return &store{m, pg, bucketName, rdb, userHistory}
}

func (s *store) GetBeatByID(ctx context.Context, beatID int64) (*core.Beat, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var beat core.Beat

	stmt := `SELECT id, beatmaker_id, path, is_downloaded, is_deleted, created_at, updated_at
	FROM beats
	WHERE id = $1`

	err := s.DB.QueryRowContext(ctx, stmt, beatID).Scan(
		&beat.ID,
		&beat.BeatmakerID,
		&beat.Path,
		&beat.IsDownloaded,
		&beat.IsDeleted,
		&beat.CreatedAt,
		&beat.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, core.ErrBeatNotFound
		}
		return nil, err
	}

	return &beat, nil
}

func (s *store) GetBeatFromS3(ctx context.Context, beatPath string, start int64, end *int64) (*miniolib.Object, int64, string, error) {
	objInfo, err := s.Minio.Client.StatObject(ctx, s.bucketName, beatPath, miniolib.StatObjectOptions{})
	if err != nil {
		return nil, 0, "", err
	}

	if start >= objInfo.Size || start < 0 {
		return nil, 0, "", core.ErrInvalidRange
	}

	if *end >= objInfo.Size || start > 0 && *end == -1 {
		*end = objInfo.Size - 1
	}

	opts := miniolib.GetObjectOptions{}
	if start != 0 || *end != -1 {
		opts.SetRange(start, *end)
	}

	obj, err := s.Minio.Client.GetObject(ctx, s.bucketName, beatPath, opts)
	if err != nil {
		return nil, 0, "", err
	}

	return obj, objInfo.Size, objInfo.ContentType, nil
}

func (s *store) AddBeat(ctx context.Context, beat core.Beat, beatGenre []core.BeatGenre) (beatID int, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	tx, err := s.DB.Begin()
	if err != nil {
		return 0, err
	}

	defer func() {
		if err != nil {
			tx.Rollback() // nolint
		} else {
			tx.Commit() // nolint
		}
	}()

	stmt := `INSERT INTO beats (id, beatmaker_id, path)
	VALUES ($1, $2, $3)
	RETURNING id`
	err = tx.QueryRowContext(
		ctx,
		stmt,
		beat.ID,
		beat.BeatmakerID,
		beat.Path).Scan(&beatID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return 0, core.ErrBeatExists
		}
		return 0, err
	}

	stmt = `INSERT INTO beats_genres (beat_id, genre)
	VALUES ($1, $2)`
	for _, bg := range beatGenre {
		err = tx.QueryRowContext(ctx, stmt, bg.BeatID, bg.Genre).Err()
		if err != nil {
			return 0, err
		}
	}

	return beatID, nil
}

func (s *store) GetPresignedURL(ctx context.Context, path string, expiry time.Duration) (url string, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	u, err := s.Minio.Client.PresignedPutObject(ctx, s.bucketName, path, expiry)
	if err != nil {
		return "", err
	}

	return u.String(), nil
}

func (s *store) GetBeatByParams(ctx context.Context, params core.BeatParams, seen []string) (beat *core.Beat, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt :=
		`WITH a AS (
			SELECT bg.id, bg.beat_id, bg.genre, b.beatmaker_id, b.is_downloaded, b.is_deleted, b.created_at FROM beats_genres bg
			JOIN beats b ON bg.beat_id = b.id
			WHERE bg.genre LIKE $1
			AND b.is_downloaded = true
			AND b.is_deleted = false
			AND bg.beat_id NOT IN (SELECT UNNEST($2::int[]))
			ORDER BY random()
		), b as (
			SELECT * FROM a
			OFFSET FLOOR(random() * (SELECT COUNT(*) FROM a))
		)

		SELECT beat_id, beatmaker_id, created_at FROM b LIMIT 1`
	beat = new(core.Beat)
	err = s.DB.QueryRowContext(
		ctx,
		stmt,
		params.Genre+"%",
		seen).Scan(
		&beat.ID,
		&beat.BeatmakerID,
		&beat.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, core.ErrBeatNotFound
		}
		return nil, err
	}

	return beat, nil
}

func (s *store) GetUserSeenBeats(ctx context.Context, userID int) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	res, err := s.Redis.Client.LRange(ctx, fmt.Sprintf("%d", userID), 0, -1).Result()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *store) AddUserSeenBeat(ctx context.Context, userID int, beatID int) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := s.Redis.Client.LPush(ctx, fmt.Sprintf("%d", userID), beatID).Result()
	if err != nil {
		return err
	}

	return nil
}

func (s *store) PopUserSeenBeat(ctx context.Context, userID int) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("%d", userID)

	if cnt, err := s.Redis.Client.LLen(ctx, key).Result(); err != nil || cnt <= int64(s.userHistory) {
		return err
	}

	_, err := s.Redis.Client.RPop(ctx, key).Result()
	if err != nil {
		return err
	}

	return nil
}

func (s *store) ClearUserSeenBeats(ctx context.Context, userID int) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if _, err := s.Redis.Client.Del(ctx, fmt.Sprintf("%d", userID)).Result(); err != nil {
		return err
	}

	return nil
}
