package beat

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/db/generated"
	miniolib "github.com/minio/minio-go/v7"

	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/minio"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/postgres"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/model"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type BeatStore struct {
	*minio.Minio
	*postgres.Postgres
	*generated.Queries
	bucketName string
}

func New(
	m *minio.Minio,
	pg *postgres.Postgres,
	bucketName string) *BeatStore {
	return &BeatStore{m, pg, generated.New(pg.DB), bucketName}
}

func (s *BeatStore) SaveBeat(ctx context.Context, beat model.SaveBeatParams) (err error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return err
	}

	defer tx.Rollback(ctx)

	qtx := s.Queries.WithTx(tx)

	if err = qtx.SaveBeat(ctx, beat.SaveBeatParams); err != nil {
		logger.Log().Error(ctx, err.Error())
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.ErrBeatAlreadyExists
		}
		return err
	}

	if _, err = qtx.SaveGenres(ctx, beat.Genres); err != nil {
		logger.Log().Error(ctx, err.Error())
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return model.ErrInvalidGenreID
		}
		return err
	}

	if _, err = qtx.SaveTags(ctx, beat.Tags); err != nil {
		logger.Log().Error(ctx, err.Error())
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return model.ErrInvalidTagID
		}
		return err
	}

	if _, err = qtx.SaveMoods(ctx, beat.Moods); err != nil {
		logger.Log().Error(ctx, err.Error())
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return model.ErrInvalidMoodID
		}
		return err
	}

	if err = qtx.SaveNote(ctx, beat.Note); err != nil {
		logger.Log().Error(ctx, err.Error())
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return model.ErrInvalidNoteID
		}
		return err
	}

	return tx.Commit(ctx)
}

func (s *BeatStore) GetBeatByID(ctx context.Context, id int) (*generated.Beat, error) {
	beat, err := s.Queries.GetBeatByID(ctx, int32(id))
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrBeatNotFound
		}
	}

	return &beat, nil
}

func (s *BeatStore) GetSaveFileURL(ctx context.Context, path string) (*string, error) {
	expiry := time.Second * 60 * 60 * 24

	url, err := s.Minio.Client.PresignedPutObject(ctx, s.bucketName, path, expiry)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	u := url.String()
	return &u, nil
}

func (s *BeatStore) GetDownloadFileURL(ctx context.Context, path string) (*string, error) {
	expiry := time.Second * 60 * 60 * 24

	url, err := s.Minio.Client.PresignedGetObject(ctx, s.bucketName, path, expiry, nil)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	u := url.String()
	return &u, nil
}

func (s *BeatStore) GetBeats(ctx context.Context, params model.GetBeatsParams) (beats []model.Beat, total *int, err error) {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query := builder.Select(
		"b.id",
		"b.beatmaker_id",
		"b.file_path",
		"b.image_path",
		"b.name",
		"b.description",
		"b.is_file_downloaded",
		"b.is_image_downloaded",
		"b.bpm",
		"b.created_at",
		"array_agg(distinct g.name) filter (where g.name is not null) as genres",
		"array_agg(distinct t.name) filter (where t.name is not null) as tags",
		"array_agg(distinct m.name) filter (where m.name is not null) as moods",
		"n.name note_name",
		"bn.scale note_scale",
	).Distinct().From("beats b").
		LeftJoin("beats_genres bg on b.id = bg.beat_id").
		LeftJoin("beats_tags bt on b.id = bt.beat_id").
		LeftJoin("beats_moods bm on b.id = bm.beat_id").
		LeftJoin("beats_notes bn on b.id = bn.beat_id").
		LeftJoin("genres g on bg.genre_id = g.id").
		LeftJoin("tags t on bt.tag_id = t.id").
		LeftJoin("moods m on bm.mood_id = m.id").
		LeftJoin("notes n on bn.note_id = n.id").
		Where("b.is_deleted = false").
		GroupBy("b.id", "n.name", "bn.scale")

	if params.BeatID != nil {
		query = query.Where("id = ?", *params.BeatID)
	}
	if params.BeatmakerID != nil {
		query = query.Where("beatmaker_id = ?", *params.BeatmakerID)
	}
	if params.BeatName != nil {
		query = query.Where("name = ?", *params.BeatName)
	}
	if params.Bpm != nil {
		query = query.Where("bpm between (?-15) and (?+15)", *params.Bpm, *params.Bpm)
	}
	if params.IsDownloaded != nil {
		if *params.IsDownloaded {
			query = query.Where("is_file_downloaded = ? and is_image_downloaded = ?", params.IsDownloaded, params.IsDownloaded)
		} else {
			query = query.Where("is_file_downloaded = ? or is_image_downloaded = ?", params.IsDownloaded, params.IsDownloaded)
		}
	}
	if params.Genre != nil {
		query = query.Where("genre in (?)", params.Genre)
	}
	if params.Tag != nil {
		query = query.Where("tag in (?)", params.Tag)
	}
	if params.Mood != nil {
		query = query.Where("mood in (?)", params.Mood)
	}
	if params.Note != nil {
		query = query.Where("note_name = ? and note_scale = ?", params.Note.Name, params.Note.Scale)
	}

	count := builder.Select("count(distinct b.id)").FromSelect(query, "b")
	sql, args, err := count.ToSql()
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	logger.Log().Debug(ctx, sql)

	if err = s.DB.QueryRow(ctx, sql, args...).Scan(&total); err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	if params.OrderBy != nil {
		query = query.OrderBy(fmt.Sprintf("%q %s", params.OrderBy.Field, params.OrderBy.Order))
	}
	query = query.Limit(uint64(params.Limit)).Offset(uint64(params.Offset))

	sql, args, err = query.ToSql()
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	rows, err := s.DB.Query(ctx, sql, args...)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42703" {
			return nil, nil, model.ErrOrderByInvalidField
		}
		return nil, nil, err
	}
	defer rows.Close()

	beats, err = pgx.CollectRows(rows, pgx.RowToStructByName[model.Beat])
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	return beats, total, nil
}

func (s *BeatStore) GetBeatParams(ctx context.Context) (params *model.BeatParams, err error) {
	genres, err := s.Queries.GetBeatGenreParams(ctx)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	moods, err := s.Queries.GetBeatMoodParams(ctx)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	tags, err := s.Queries.GetBeatTagParams(ctx)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	notes, err := s.Queries.GetBeatNoteParams(ctx)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	return &model.BeatParams{
		Genres: genres,
		Moods:  moods,
		Tags:   tags,
		Notes:  notes,
	}, nil
}

func (s *BeatStore) GetBeatBytes(ctx context.Context, path string, start, end *int) (file io.ReadCloser, size *int, contentType *string, err error) {
	info, err := s.Minio.Client.StatObject(ctx, s.bucketName, path, miniolib.StatObjectOptions{})
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, nil, model.ErrBeatNotFound
	}

	tmp := int(info.Size)
	size = &tmp

	if start != nil && *start >= *size {
		logger.Log().Error(ctx, model.ErrInvalidRangeHeader.Error())
		return nil, nil, nil, model.ErrInvalidRangeHeader
	}

	if end != nil && *end >= *size || start != nil && end == nil {
		end = new(int)
		*end = int(*size - 1)
	}

	var opts miniolib.GetObjectOptions
	if start != nil {
		if err := opts.SetRange(int64(*start), int64(*end)); err != nil {
			logger.Log().Error(ctx, err.Error())
			return nil, nil, nil, err
		}
	}

	file, err = s.Minio.Client.GetObject(ctx, s.bucketName, path, opts)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, nil, err
	}

	return file, size, &info.ContentType, nil
}

func (s *BeatStore) UpdateBeat(ctx context.Context, updateBeat model.UpdateBeatParams) (*generated.Beat, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	defer tx.Rollback(ctx)

	qtx := s.Queries.WithTx(tx)
	beat := new(generated.Beat)
	if *beat, err = qtx.UpdateBeat(ctx, updateBeat.UpdateBeatParams); err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrBeatNotFound
		}
		return nil, err
	}

	if updateBeat.Genres != nil {
		if err = qtx.DeleteBeatGenres(ctx, updateBeat.ID); err != nil {
			logger.Log().Error(ctx, err.Error())
			return nil, err
		}

		if _, err = qtx.SaveGenres(ctx, updateBeat.Genres); err != nil {
			logger.Log().Error(ctx, err.Error())
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				return nil, model.ErrInvalidGenreID
			}
			return nil, err
		}
	}

	if updateBeat.Tags != nil {
		if err = qtx.DeleteBeatTags(ctx, updateBeat.ID); err != nil {
			logger.Log().Error(ctx, err.Error())
			return nil, err
		}

		if _, err = qtx.SaveTags(ctx, updateBeat.Tags); err != nil {
			logger.Log().Error(ctx, err.Error())
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				return nil, model.ErrInvalidTagID
			}
			return nil, err
		}
	}

	if updateBeat.Moods != nil {
		if err = qtx.DeleteBeatMoods(ctx, updateBeat.ID); err != nil {
			logger.Log().Error(ctx, err.Error())
			return nil, err
		}

		if _, err = qtx.SaveMoods(ctx, updateBeat.Moods); err != nil {
			logger.Log().Error(ctx, err.Error())
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				return nil, model.ErrInvalidMoodID
			}
			return nil, err
		}
	}

	if updateBeat.Note != nil {
		if err = qtx.DeleteBeatNotes(ctx, updateBeat.ID); err != nil {
			logger.Log().Error(ctx, err.Error())
			return nil, err
		}

		if err = qtx.SaveNote(ctx, *updateBeat.Note); err != nil {
			logger.Log().Error(ctx, err.Error())
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				return nil, model.ErrInvalidNoteID
			}
			return nil, err
		}
	}

	return beat, tx.Commit(ctx)
}

func (s *BeatStore) DeleteBeat(ctx context.Context, id int) error {
	if err := s.Queries.DeleteBeat(ctx, int32(id)); err != nil {
		logger.Log().Error(ctx, err.Error())
		return err
	}

	return nil
}
