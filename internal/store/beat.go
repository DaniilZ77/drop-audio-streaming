package beat

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
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

func (s *store) GetBeatByID(ctx context.Context, beatID int, param core.IsDownloaded) (*core.Beat, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var beat core.Beat

	isDownloaded := []bool{true}
	if param == core.False {
		isDownloaded = []bool{false}
	} else if param == core.Any {
		isDownloaded = append(isDownloaded, false)
	}

	logger.Log().Debug(ctx, "%v", isDownloaded)

	stmt := `SELECT id, beatmaker_id, file_path, image_path, name, description, is_file_downloaded, is_image_downloaded, is_deleted, created_at, updated_at
	FROM beats
	WHERE id = $1 AND is_file_downloaded = ANY($2::boolean[]) AND is_image_downloaded = ANY($2::boolean[]) AND is_deleted = false`
	err := s.DB.QueryRowContext(ctx, stmt, beatID, isDownloaded).Scan(
		&beat.ID,
		&beat.BeatmakerID,
		&beat.FilePath,
		&beat.ImagePath,
		&beat.Name,
		&beat.Description,
		&beat.IsFileDownloaded,
		&beat.IsImageDownloaded,
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

func (s *store) GetBeatFromS3(ctx context.Context, beatPath string, start int, end *int) (obj io.ReadCloser, size int, contentType string, err error) {
	objInfo, err := s.Minio.Client.StatObject(ctx, s.bucketName, beatPath, miniolib.StatObjectOptions{})
	if err != nil {
		return nil, 0, "", err
	}

	size = int(objInfo.Size)

	if start >= size || start < 0 {
		return nil, 0, "", core.ErrInvalidRange
	}

	if *end >= size || start > 0 && *end == -1 {
		*end = size - 1
	}

	opts := miniolib.GetObjectOptions{}
	if start != 0 || *end != -1 {
		if err := opts.SetRange(int64(start), int64(*end)); err != nil {
			return nil, 0, "", err
		}
	}

	obj, err = s.Minio.Client.GetObject(ctx, s.bucketName, beatPath, opts)
	if err != nil {
		return nil, 0, "", err
	}

	return obj, size, objInfo.ContentType, nil
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

	stmt := `INSERT INTO beats (id, beatmaker_id, file_path, image_path, name, description)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id`
	err = tx.QueryRowContext(
		ctx,
		stmt,
		beat.ID,
		beat.BeatmakerID,
		beat.FilePath,
		beat.ImagePath,
		beat.Name,
		beat.Description).Scan(&beatID)
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
		err = tx.QueryRowContext(ctx, stmt, beatID, bg.Genre).Err()
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

func (s *store) GetBeatByFilter(ctx context.Context, filter core.FeedFilter, seen []string) (beat *core.Beat, genre *string, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt :=
		`WITH a AS (
			SELECT bg.id, bg.beat_id, bg.genre, b.beatmaker_id, b.image_path, b.name, b.description, b.created_at FROM beats_genres bg
			JOIN beats b ON bg.beat_id = b.id
			WHERE bg.genre LIKE $1
			AND b.is_file_downloaded = true
			AND b.is_image_downloaded
			AND b.is_deleted = false
			AND bg.beat_id NOT IN (SELECT UNNEST($2::int[]))
		), b as (
			SELECT * FROM a
			OFFSET FLOOR(random() * (SELECT COUNT(*) FROM a))
		)

		SELECT beat_id, beatmaker_id, image_path, name, description, created_at, genre FROM b LIMIT 1`
	beat = new(core.Beat)
	genre = new(string)
	err = s.DB.QueryRowContext(
		ctx,
		stmt,
		filter.Genre+"%",
		seen).Scan(
		&beat.ID,
		&beat.BeatmakerID,
		&beat.ImagePath,
		&beat.Name,
		&beat.Description,
		&beat.CreatedAt,
		genre,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, core.ErrBeatNotFound
		}
		return nil, nil, err
	}

	return beat, genre, nil
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

func (s *store) AddUserSeenBeat(ctx context.Context, userID, beatID int) error {
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

func (s *store) GetBeatGenres(ctx context.Context, beatID int) (beatGenres []core.BeatGenre, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt := `SELECT id, beat_id, genre
	FROM beats_genres
	WHERE beat_id = $1`

	rows, err := s.DB.QueryContext(ctx, stmt, beatID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var beatGenre core.BeatGenre
		if err := rows.Scan(&beatGenre.ID, &beatGenre.BeatID, &beatGenre.Genre); err != nil {
			return nil, err
		}
		beatGenres = append(beatGenres, beatGenre)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return beatGenres, nil
}

func (s *store) GetBeatsByBeatmakerID(ctx context.Context, beatmakerID int, p core.GetBeatsParams) (beats []core.Beat, total int, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, 0, err
	}

	defer func() {
		if err != nil {
			tx.Rollback() // nolint
		} else {
			tx.Commit() // nolint
		}
	}()

	stmt := `SELECT COUNT(*)
	FROM beats
	WHERE beatmaker_id = $1 AND is_file_downloaded = true AND is_image_downloaded = true AND is_deleted = false`

	err = s.DB.QueryRowContext(ctx, stmt, beatmakerID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	stmt = fmt.Sprintf(`SELECT id, beatmaker_id, file_path, image_path, name, description, is_file_downloaded, is_image_downloaded, is_deleted, created_at, updated_at
	FROM beats
	WHERE beatmaker_id = $1 AND is_deleted = false AND is_file_downloaded = true AND is_image_downloaded = true
	ORDER BY updated_at %s
	OFFSET $2
	LIMIT $3`, p.Order)

	logger.Log().Debug(ctx, "order: %s", p.Order)

	rows, err := s.DB.QueryContext(ctx, stmt, beatmakerID, p.Offset, p.Limit)
	if err != nil {
		return nil, 0, err
	}

	for rows.Next() {
		var beat core.Beat
		if err := rows.Scan(
			&beat.ID,
			&beat.BeatmakerID,
			&beat.FilePath,
			&beat.ImagePath,
			&beat.Name,
			&beat.Description,
			&beat.IsFileDownloaded,
			&beat.IsImageDownloaded,
			&beat.IsDeleted,
			&beat.CreatedAt,
			&beat.UpdatedAt); err != nil {
			return nil, 0, err
		}
		beats = append(beats, beat)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, nil
	}

	return beats, total, nil
}
