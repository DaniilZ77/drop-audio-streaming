package beat

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/db/generated"
	miniolib "github.com/minio/minio-go/v7"

	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/domain/model"
	sl "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/minio"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/postgres"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type BeatStore struct {
	*minio.Minio
	*postgres.Postgres
	*generated.Queries
	bucketName string
	log        *slog.Logger
}

func New(
	m *minio.Minio,
	pg *postgres.Postgres,
	bucketName string,
	log *slog.Logger) *BeatStore {
	return &BeatStore{m, pg, generated.New(pg.DB), bucketName, log}
}

func (s *BeatStore) SaveBeat(ctx context.Context, beat model.SaveBeat) (err error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		s.log.Error("failed to start transaction", sl.Err(err))
		return err
	}

	defer tx.Rollback(ctx) // nolint

	qtx := s.Queries.WithTx(tx)

	if err = qtx.SaveBeat(ctx, beat.SaveBeatParams); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return &model.ModelError{Err: model.ErrBeatAlreadyExists}
		}
		s.log.Error("failed to save beat", sl.Err(err))
		return err
	}

	if _, err = qtx.SaveGenres(ctx, beat.Genres); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return model.NewErr(model.ErrInvalidID, "genre id")
		}
		s.log.Error("failed to save genres", sl.Err(err))
		return err
	}

	if _, err = qtx.SaveTags(ctx, beat.Tags); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return model.NewErr(model.ErrInvalidID, "tag id")
		}
		s.log.Error("failed to save tags", sl.Err(err))
		return err
	}

	if _, err = qtx.SaveMoods(ctx, beat.Moods); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return model.NewErr(model.ErrInvalidID, "mood id")
		}
		s.log.Error("failed to save moods", sl.Err(err))
		return err
	}

	if err = qtx.SaveNote(ctx, beat.Note); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return model.NewErr(model.ErrInvalidID, "note id")
		}
		s.log.Error("failed to save note", sl.Err(err))
		return err
	}

	return tx.Commit(ctx)
}

func (s *BeatStore) GetBeatByID(ctx context.Context, id uuid.UUID) (*generated.Beat, error) {
	beat, err := s.Queries.GetBeatByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &model.ModelError{Err: model.ErrBeatNotFound}
		}
		return nil, err
	}

	return &beat, nil
}

func (s *BeatStore) GetDownloadMediaURL(ctx context.Context, path string, expires time.Duration) (*string, error) {
	url, err := s.Minio.Client.PresignedGetObject(ctx, s.bucketName, path, expires, nil)
	if err != nil {
		return nil, err
	}

	u := url.RequestURI()
	return &u, nil
}

func (s *BeatStore) GetBeats(ctx context.Context, params model.GetBeatsParams) (beats []model.Beat, total *uint64, err error) {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query := builder.Select(
		"b.id",
		"b.beatmaker_id",
		"b.image_path",
		"b.name",
		"b.description",
		"b.is_file_downloaded",
		"b.is_image_downloaded",
		"b.is_archive_downloaded",
		"b.bpm",
		"b.range_start",
		"b.range_end",
		"b.created_at",
		"array_agg(distinct g.name) filter (where g.name is not null) as genres",
		"array_agg(distinct t.name) filter (where t.name is not null) as tags",
		"array_agg(distinct m.name) filter (where m.name is not null) as moods",
		"n.name note_name",
		"bn.scale note_scale",
	).From("beats b").
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
		query = query.Where("b.id = ?", *params.BeatID)
	}
	if params.BeatmakerID != nil {
		query = query.Where("b.beatmaker_id = ?", *params.BeatmakerID)
	}
	if params.BeatName != nil {
		query = query.Where("b.name = ?", *params.BeatName)
	}
	if params.Bpm != nil {
		query = query.Where("b.bpm between (?-15) and (?+15)", *params.Bpm, *params.Bpm)
	}
	if params.IsDownloaded != nil {
		if *params.IsDownloaded {
			query = query.Where("b.is_file_downloaded = ? and b.is_image_downloaded = ? and b.is_archive_downloaded = ?", params.IsDownloaded, params.IsDownloaded, params.IsDownloaded)
		} else {
			query = query.Where("(b.is_file_downloaded = ? or b.is_image_downloaded = ? or b.is_archive_downloaded = ?)", params.IsDownloaded, params.IsDownloaded, params.IsDownloaded)
		}
	}

	if params.Genre != nil {
		query = query.Where("g.name = any(?)", params.Genre)
	}
	if params.Tag != nil {
		query = query.Where("t.name = any(?)", params.Tag)
	}
	if params.Mood != nil {
		query = query.Where("m.name = any(?)", params.Mood)
	}
	if params.Note != nil {
		query = query.Where("n.name = ? and bn.scale = ?", params.Note.Name, params.Note.Scale)
	}

	count := builder.Select("count(distinct b.id)").FromSelect(query, "b")
	sql, args, err := count.ToSql()
	if err != nil {
		s.log.Error("failed to convert to sql", sl.Err(err))
		return nil, nil, err
	}

	if err = s.DB.QueryRow(ctx, sql, args...).Scan(&total); err != nil {
		s.log.Error("failed to count beats", sl.Err(err))
		return nil, nil, err
	}

	if params.OrderBy != nil {
		query = query.OrderBy(fmt.Sprintf("%q %s", params.OrderBy.Field, params.OrderBy.Order))
	}
	query = query.Limit(params.Limit).Offset(params.Offset)

	sql, args, err = query.ToSql()
	if err != nil {
		s.log.Error("failed to convert to sql", sl.Err(err))
		return nil, nil, err
	}

	rows, err := s.DB.Query(ctx, sql, args...)
	if err != nil {
		s.log.Error("failed to get beats", sl.Err(err))
		return nil, nil, err
	}
	defer rows.Close()

	beats, err = pgx.CollectRows(rows, pgx.RowToStructByName[model.Beat])
	if err != nil {
		s.log.Error("failed to collect beats", sl.Err(err))
		return nil, nil, err
	}

	return beats, total, nil
}

func (s *BeatStore) GetBeatParams(ctx context.Context) (attrs *model.BeatAttributes, err error) {
	genres, err := s.Queries.GetBeatGenreParams(ctx)
	if err != nil {
		s.log.Error("failed to get beat genre params", sl.Err(err))
		return nil, err
	}

	moods, err := s.Queries.GetBeatMoodParams(ctx)
	if err != nil {
		s.log.Error("failed to get beat mood params", sl.Err(err))
		return nil, err
	}

	tags, err := s.Queries.GetBeatTagParams(ctx)
	if err != nil {
		s.log.Error("failed to get beat tag params", sl.Err(err))
		return nil, err
	}

	notes, err := s.Queries.GetBeatNoteParams(ctx)
	if err != nil {
		s.log.Error("failed to get beat note params", sl.Err(err))
		return nil, err
	}

	return &model.BeatAttributes{
		Genres: genres,
		Moods:  moods,
		Tags:   tags,
		Notes:  notes,
	}, nil
}

func (s *BeatStore) GetBeatBytes(ctx context.Context, path string, start, end *int) (file io.ReadCloser, size *int, contentType *string, err error) {
	info, err := s.Minio.Client.StatObject(ctx, s.bucketName, path, miniolib.StatObjectOptions{})
	if err != nil {
		s.log.Error("failed to get beat info", sl.Err(err))
		return nil, nil, nil, err
	}

	tmp := int(info.Size)
	size = &tmp

	if start != nil && *start >= *size {
		s.log.Debug("start is greater than size", slog.Int("start", *start), slog.Int("size", *size))
		return nil, nil, nil, &model.ModelError{Err: model.ErrInvalidRangeHeader}
	}

	if end != nil && *end >= *size || start != nil && end == nil {
		end = new(int)
		*end = int(*size - 1)
	}

	var opts miniolib.GetObjectOptions
	if start != nil {
		if err := opts.SetRange(int64(*start), int64(*end)); err != nil {
			s.log.Error("failed to set range", sl.Err(err))
			return nil, nil, nil, err
		}
	}

	file, err = s.Minio.Client.GetObject(ctx, s.bucketName, path, opts)
	if err != nil {
		s.log.Error("failed to get beat", sl.Err(err))
		return nil, nil, nil, err
	}

	return file, size, &info.ContentType, nil
}

func (s *BeatStore) UpdateBeat(ctx context.Context, updateBeat model.UpdateBeat) (*generated.Beat, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		s.log.Error("failed to start transaction", sl.Err(err))
		return nil, err
	}

	defer tx.Rollback(ctx) // nolint

	qtx := s.Queries.WithTx(tx)
	beat := new(generated.Beat)
	if *beat, err = qtx.UpdateBeat(ctx, updateBeat.UpdateBeatParams); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &model.ModelError{Err: model.ErrBeatNotFound}
		}
		s.log.Error("failed to update beat", sl.Err(err))
		return nil, err
	}

	if updateBeat.Genres != nil {
		if err = qtx.DeleteBeatGenres(ctx, updateBeat.ID); err != nil {
			s.log.Error("failed to delete beat genres", sl.Err(err))
			return nil, err
		}

		if _, err = qtx.SaveGenres(ctx, updateBeat.Genres); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				return nil, model.NewErr(model.ErrInvalidID, "genre id")
			}
			s.log.Error("failed to save genres", sl.Err(err))
			return nil, err
		}
	}

	if updateBeat.Tags != nil {
		if err = qtx.DeleteBeatTags(ctx, updateBeat.ID); err != nil {
			s.log.Error("failed to delete beat tags", sl.Err(err))
			return nil, err
		}

		if _, err = qtx.SaveTags(ctx, updateBeat.Tags); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				return nil, model.NewErr(model.ErrInvalidID, "tag id")
			}
			s.log.Error("failed to save tags", sl.Err(err))
			return nil, err
		}
	}

	if updateBeat.Moods != nil {
		if err = qtx.DeleteBeatMoods(ctx, updateBeat.ID); err != nil {
			s.log.Error("failed to delete beat moods", sl.Err(err))
			return nil, err
		}

		if _, err = qtx.SaveMoods(ctx, updateBeat.Moods); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				return nil, model.NewErr(model.ErrInvalidID, "mood id")
			}
			s.log.Error("failed to save moods", sl.Err(err))
			return nil, err
		}
	}

	if updateBeat.Note != nil {
		if err = qtx.DeleteBeatNotes(ctx, updateBeat.ID); err != nil {
			s.log.Error("failed to delete beat notes", sl.Err(err))
			return nil, err
		}

		if err = qtx.SaveNote(ctx, *updateBeat.Note); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				return nil, model.NewErr(model.ErrInvalidID, "note id")
			}
			s.log.Error("failed to save note", sl.Err(err))
			return nil, err
		}
	}

	return beat, tx.Commit(ctx)
}

func (s *BeatStore) DeleteBeat(ctx context.Context, id uuid.UUID) error {
	if err := s.Queries.DeleteBeat(ctx, id); err != nil {
		return err
	}

	return nil
}

func (s *BeatStore) UploadMedia(ctx context.Context, path, contentType string, file io.Reader) error {
	if _, err := s.Minio.Client.PutObject(ctx, s.bucketName, path, file, -1, miniolib.PutObjectOptions{ContentType: contentType}); err != nil {
		return err
	}

	return nil
}

func (s *BeatStore) SaveOwner(ctx context.Context, owner generated.SaveOwnerParams) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		s.log.Error("failed to start transaction", sl.Err(err))
		return err
	}

	defer tx.Rollback(ctx) // nolint

	qtx := s.Queries.WithTx(tx)
	if err := qtx.DeleteBeat(ctx, owner.BeatID); err != nil {
		s.log.Error("failed to delete beat", sl.Err(err))
		return err
	}

	if err := qtx.SaveOwner(ctx, owner); err != nil {
		s.log.Error("failed to save owner", sl.Err(err))
		return err
	}

	return tx.Commit(ctx)
}

func (s *BeatStore) GetOwnerByBeatID(ctx context.Context, beatID uuid.UUID) (*generated.BeatsOwner, error) {
	owner, err := s.Queries.GetOwnerByBeatID(ctx, beatID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &model.ModelError{Err: model.ErrOwnerNotFound}
		}
		return nil, err
	}

	return &owner, err
}
