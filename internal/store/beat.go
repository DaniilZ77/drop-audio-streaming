package beat

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/minio"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/postgres"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/redis"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	miniolib "github.com/minio/minio-go/v7"
	redislib "github.com/redis/go-redis/v9"
)

const (
	redisRetries = 7
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

func getFullBeatByID(ctx context.Context, tx *sql.Tx, beatID int, beat core.Beat) (*core.BeatParams, error) {
	var tags []core.BeatTag
	stmt := `SELECT tag_id FROM beats_tags WHERE beat_id = $1`
	res, err := tx.QueryContext(ctx, stmt, beatID)
	for res.Next() {
		var tag core.BeatTag
		err = res.Scan(&tag.TagID)
		if err != nil {
			return nil, err
		}

		tags = append(tags, tag)
	}

	var moods []core.BeatMood
	stmt = `SELECT mood_id FROM beats_moods WHERE beat_id = $1`
	res, err = tx.QueryContext(ctx, stmt, beatID)
	for res.Next() {
		var mood core.BeatMood
		err = res.Scan(&mood.MoodID)
		if err != nil {
			return nil, err
		}

		moods = append(moods, mood)
	}

	var genres []core.BeatGenre
	stmt = `SELECT genre_id FROM beats_genres WHERE beat_id = $1`
	res, err = tx.QueryContext(ctx, stmt, beatID)
	for res.Next() {
		var genre core.BeatGenre
		err = res.Scan(&genre.GenreID)
		if err != nil {
			return nil, err
		}

		genres = append(genres, genre)
	}

	var note core.BeatNote
	stmt = `SELECT note_id, scale FROM beats_notes WHERE beat_id = $1`
	err = tx.QueryRowContext(ctx, stmt, beatID).Scan(&note.NoteID, &note.Scale)
	if err != nil {
		return nil, err
	}

	return &core.BeatParams{
		Beat:   beat,
		Tags:   tags,
		Moods:  moods,
		Genres: genres,
		Note:   note,
	}, nil
}

func (s *store) GetFullBeatByID(ctx context.Context, id int, param core.IsDownloaded) (*core.BeatParams, error) {
	stmt := `SELECT id, beatmaker_id, file_path, image_path, name, bpm, description,
	is_file_downloaded, is_image_downloaded, is_deleted, created_at, updated_at
	FROM beats
	WHERE id = $1
	AND is_deleted = false
	AND is_file_downloaded = true
	AND is_image_downloaded = true`

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback() // nolint
		} else {
			tx.Commit() // nolint
		}
	}()

	var beat core.Beat
	err = tx.QueryRowContext(ctx, stmt, id).Scan(
		&beat.ID,
		&beat.BeatmakerID,
		&beat.FilePath,
		&beat.ImagePath,
		&beat.Name,
		&beat.Bpm,
		&beat.Description,
		&beat.IsFileDownloaded,
		&beat.IsImageDownloaded,
		&beat.IsDeleted,
		&beat.CreatedAt,
		&beat.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, core.ErrBeatNotFound
		}
		return nil, err
	}

	beatParams, err := getFullBeatByID(ctx, tx, id, beat)
	if err != nil {
		return nil, err
	}

	return beatParams, nil
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

	stmt := `SELECT id, beatmaker_id, file_path, image_path, name,description,
	is_file_downloaded, is_image_downloaded, is_deleted, created_at, updated_at
	FROM beats
	WHERE id = $1
	AND is_file_downloaded = ANY($2::boolean[])
	AND is_image_downloaded = ANY($2::boolean[])
	AND is_deleted = false`
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

	var opts miniolib.GetObjectOptions
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

func insertTx(ctx context.Context, stmt string, beatID int, elems []int, tx *sql.Tx) error {
	var args []any
	args = append(args, beatID)
	stmt = `INSERT INTO beats_genres (beat_id, genre_id)
	VALUES`
	cur := 1
	for _, elem := range elems {
		stmt += fmt.Sprintf(" ($%d, $%d),", cur, cur+1)
		args = append(args, elem)
		cur += 2
	}
	stmt = strings.TrimSuffix(stmt, ",")

	_, err := tx.ExecContext(ctx, stmt, args...)
	if err != nil {
		return err
	}

	return nil
}

func (s *store) AddBeat(ctx context.Context, beat core.BeatParams) (beatID int, err error) {
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
		beat.Beat.ID,
		beat.Beat.BeatmakerID,
		beat.Beat.FilePath,
		beat.Beat.ImagePath,
		beat.Beat.Name,
		beat.Beat.Description).Scan(&beatID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return 0, core.ErrBeatExists
		}
		return 0, err
	}

	var genres []int
	for _, genre := range beat.Genres {
		genres = append(genres, genre.GenreID)
	}
	if err := insertTx(ctx, `INSERT INTO beats_genres (beat_id, genre_id)
	VALUES`, beatID, genres, tx); err != nil {
		return 0, err
	}

	var tags []int
	for _, tag := range beat.Tags {
		tags = append(tags, tag.TagID)
	}
	if err := insertTx(ctx, `INSERT INTO beats_tags (beat_id, tag_id)
	VALUES`, beatID, tags, tx); err != nil {
		return 0, err
	}

	var moods []int
	for _, mood := range beat.Moods {
		moods = append(moods, mood.MoodID)
	}
	if err := insertTx(ctx, `INSERT INTO beats_moods (beat_id, mood)
	VALUES`, beatID, moods, tx); err != nil {
		return 0, err
	}

	stmt = `INSERT INTO beats_notes (beat_id, note_id, scale)
	VALUES ($1, $2, $3)`
	_, err = tx.ExecContext(ctx, stmt, beat.Note.BeatID, beat.Note.NoteID, beat.Note.Scale)
	if err != nil {
		return 0, err
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

func (s *store) GetBeatByFilter(ctx context.Context, filter core.FeedFilter, seen []string) (*core.BeatParams, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var clause string
	var args []any
	args = append(args, seen)
	cur := 2
	if filter.Genres != nil {
		clause += fmt.Sprintf("AND bg.genre_id IN (SELECT UNNEST($%d::int[]))\n", cur)
		args = append(args, filter.Genres)
		cur++
	}
	if filter.Tags != nil {
		clause += fmt.Sprintf("AND bt.tag_id IN (SELECT UNNEST($%d::int[]))\n", cur)
		args = append(args, filter.Tags)
		cur++
	}
	if filter.Moods != nil {
		clause += fmt.Sprintf("AND bm.mood_id IN (SELECT UNNEST($%d::int[]))\n", cur)
		args = append(args, filter.Moods)
		cur++
	}
	if filter.Note != nil {
		clause += fmt.Sprintf("AND bn.note_id = $%d AND bn.scale = $%d\n", cur, cur+1)
		args = append(args, filter.Note.NoteID, filter.Note.Scale)
		cur += 2
	}
	if filter.Bpm != nil {
		clause += fmt.Sprintf("AND b.bpm BETWEEN ($%d-15) AND ($%d+15)\n", cur)
		args = append(args, filter.Bpm)
		cur++
	}

	stmt := fmt.Sprintf(
		`WITH a AS (
			SELECT bg.beat_id, b.beatmaker_id, b.image_path, b.bpm, b.name, b.description, bn.note_id, bn.scale, b.created_at, b.updated_at
			FROM beats b
			JOIN beats_genres bg ON b.id = bg.beat_id
			JOIN beats_tags bt ON b.id = bt.beat_id
			JOIN beats_moods bm ON b.id = bm.beat_id
			JOIN beats_notes bn ON b.id = bn.beat_id
			WHERE bg.beat_id NOT IN (SELECT UNNEST($1::int[]))
			%s
			AND b.is_file_downloaded = true
			AND b.is_image_downloaded = true
			AND b.is_deleted = false
		), b as (
			SELECT * FROM a
			OFFSET FLOOR(random() * (SELECT COUNT(*) FROM a))
		)

		SELECT beat_id, beatmaker_id, image_path, name, description, note_id, scale, bpm, created_at, updated_at
		FROM b LIMIT 1`, clause)

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback() // nolint
		} else {
			tx.Commit() // nolint
		}
	}()

	var note core.BeatNote
	var beat core.Beat
	err = tx.QueryRowContext(
		ctx,
		stmt,
		args...).Scan(
		&beat.ID,
		&beat.BeatmakerID,
		&beat.ImagePath,
		&beat.Name,
		&beat.Description,
		&note.NoteID,
		&note.Scale,
		&beat.Bpm,
		&beat.CreatedAt,
		&beat.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, core.ErrBeatNotFound
		}
		return nil, err
	}

	beatParams, err := getFullBeatByID(ctx, tx, beat.ID, beat)
	if err != nil {
		return nil, err
	}

	return beatParams, nil
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

func (s *store) ReplaceUserSeenBeat(ctx context.Context, userID, beatID int) error {
	key := fmt.Sprintf("%d", userID)

	for range redisRetries {
		err := s.Redis.Client.Watch(ctx, func(tx *redislib.Tx) error {
			cnt, err := tx.LLen(ctx, key).Result()
			if err != nil {
				return err
			}

			_, err = tx.TxPipelined(ctx, func(pipe redislib.Pipeliner) error {
				if cnt >= int64(s.userHistory) {
					if _, err = pipe.RPop(ctx, key).Result(); err != nil {
						return err
					}
				}

				if _, err = pipe.LPush(ctx, key, beatID).Result(); err != nil {
					return err
				}

				return nil
			})
			return err
		}, key)
		if !errors.Is(err, redislib.TxFailedErr) {
			return err
		}
	}

	return core.ErrAmountOfRetriesExceeded
}

func (s *store) ClearUserSeenBeats(ctx context.Context, userID int) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if _, err := s.Redis.Client.Del(ctx, fmt.Sprintf("%d", userID)).Result(); err != nil {
		return err
	}

	return nil
}

func (s *store) GetBeatsByBeatmakerID(ctx context.Context, beatmakerID int, p core.GetBeatsParams) (beats []core.BeatParams, total int, err error) {
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
	WHERE beatmaker_id = $1
	AND is_file_downloaded = true
	AND is_image_downloaded = true
	AND is_deleted = false`

	err = s.DB.QueryRowContext(ctx, stmt, beatmakerID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	stmt = fmt.Sprintf(
		`SELECT id, beatmaker_id, file_path, image_path, name, bpm, description,
	is_file_downloaded, is_image_downloaded, is_deleted, created_at, updated_at
	FROM beats
	WHERE beatmaker_id = $1
	AND is_deleted = false
	AND is_file_downloaded = true
	AND is_image_downloaded = true
	ORDER BY updated_at %s
	OFFSET $2
	LIMIT $3`, p.Order)

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
			&beat.Bpm,
			&beat.Description,
			&beat.IsFileDownloaded,
			&beat.IsImageDownloaded,
			&beat.IsDeleted,
			&beat.CreatedAt,
			&beat.UpdatedAt); err != nil {
			return nil, 0, err
		}

		beatParams, err := getFullBeatByID(ctx, tx, beat.ID, beat)
		if err != nil {
			return nil, 0, err
		}

		beats = append(beats, *beatParams)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, nil
	}

	return beats, total, nil
}
