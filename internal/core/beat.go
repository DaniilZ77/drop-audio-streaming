package core

import (
	"context"
	"io"
	"time"
)

type (
	BeatStorage interface {
		GetBeatFromS3(ctx context.Context, beatPath string, start int, end *int) (obj io.ReadCloser, size int, contentType string, err error)
		GetFullBeatByID(ctx context.Context, id int, param IsDownloaded) (*BeatParams, error)
		GetBeatByID(ctx context.Context, id int, param IsDownloaded) (*Beat, error)
		AddBeat(ctx context.Context, beat BeatParams) (beatID int, err error)
		GetPresignedURLPut(ctx context.Context, path string, expiry time.Duration) (url string, err error)
		GetPresignedURLGet(ctx context.Context, path string, expiry time.Duration) (url string, err error)
		GetBeatByFilter(ctx context.Context, filter FeedFilter, seen []string) (beat *BeatParams, err error)
		GetBeatsByBeatmakerID(ctx context.Context, beatmakerID int, p GetBeatsParams) (beats []BeatParams, total int, err error)

		GetUserSeenBeats(ctx context.Context, userID int) ([]string, error)
		ReplaceUserSeenBeat(ctx context.Context, userID int, beatID int) error
		ClearUserSeenBeats(ctx context.Context, userID int) error
		GetFilters(ctx context.Context) (*Filters, error)
	}

	BeatService interface {
		GetBeatFromS3(ctx context.Context, beatID int, start int, end *int) (obj io.ReadCloser, size int, contentType string, err error)
		AddBeat(ctx context.Context, beat BeatParams) (beatPath, imagePath string, err error)
		WritePartialContent(ctx context.Context, r io.Reader, w io.Writer, chunkSize int) error
		GetUploadURL(ctx context.Context, beatPath string) (url string, err error)
		GetBeatByFilter(ctx context.Context, userID int, params FeedFilter) (beat *BeatParams, err error)
		GetBeat(ctx context.Context, beatID int) (beat *BeatParams, err error)
		GetBeatsByBeatmakerID(ctx context.Context, beatmakerID int, p GetBeatsParams) (beats []BeatParams, total int, err error)
		GetFilters(ctx context.Context) (*Filters, error)
		GetDownloadURL(ctx context.Context, imagePath string) (url string, err error)
	}

	FeedFilter struct {
		Genres []int
		Moods  []int
		Tags   []int
		Note   *struct {
			NoteID int
			Scale  string
		}
		Bpm *int
	}

	BeatNote struct {
		ID     int
		BeatID int
		NoteID int
		Scale  string
	}

	BeatMood struct {
		ID     int
		BeatID int
		MoodID int
	}

	BeatTag struct {
		ID     int
		BeatID int
		TagID  int
	}

	BeatGenre struct {
		ID      int
		BeatID  int
		GenreID int
	}

	Mood struct {
		ID   int
		Name string
	}

	Tag struct {
		ID   int
		Name string
	}

	Note struct {
		ID   int
		Name string
	}

	Genre struct {
		ID   int
		Name string
	}

	Filters struct {
		Moods  []Mood
		Tags   []Tag
		Genres []Genre
		Note   []Note
	}

	Beat struct {
		ID                int
		BeatmakerID       int
		FilePath          string
		ImagePath         string
		Name              string
		Description       string
		IsFileDownloaded  bool
		IsImageDownloaded bool
		IsDeleted         bool
		Bpm               int
		CreatedAt         time.Time
		UpdatedAt         time.Time
	}

	BeatParams struct {
		Beat   Beat
		Note   BeatNote
		Genres []BeatGenre
		Tags   []BeatTag
		Moods  []BeatMood
	}

	GetBeatsParams struct {
		Limit  int
		Offset int
		Order  string
	}

	IsDownloaded int
)

const (
	True IsDownloaded = iota
	False
	Any
)
