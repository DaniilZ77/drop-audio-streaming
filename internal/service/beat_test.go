package beat

import (
	"context"
	"errors"
	"io"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/db/generated"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger/slogdiscard"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/model"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/service/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type dependencies struct {
	beatService       *BeatService
	beatModifier      *mocks.BeatModifier
	beatProvider      *mocks.BeatProvider
	urlProvider       *mocks.URLProvider
	mediaUploader     *mocks.MediaUploader
	beatBytesProvider *mocks.BeatBytesProvider
	config            *BeatServiceConfig
}

func createService(t *testing.T) dependencies {
	t.Helper()

	beatModifier := mocks.NewBeatModifier(t)
	beatProvider := mocks.NewBeatProvider(t)
	urlProvider := mocks.NewURLProvider(t)
	mediaUploader := mocks.NewMediaUploader(t)
	beatBytesProvider := mocks.NewBeatBytesProvider(t)
	config := NewBeatServiceConfig(100, 200, 300, "secret", 100)

	beatService := NewBeatService(
		beatModifier,
		beatProvider,
		urlProvider,
		mediaUploader,
		beatBytesProvider,
		config,
		slogdiscard.NewDiscardLogger(),
	)

	return dependencies{
		beatService:       beatService,
		beatModifier:      beatModifier,
		beatProvider:      beatProvider,
		urlProvider:       urlProvider,
		mediaUploader:     mediaUploader,
		beatBytesProvider: beatBytesProvider,
		config:            config,
	}
}

var (
	name        = "imagine dragons"
	contentType = "audio/mp3"
	file        = strings.NewReader("content")
)

func TestUploadMedia_Success(t *testing.T) {
	t.Parallel()

	s := createService(t)

	ctx := context.Background()
	contentLength := int64(10)
	expiry := time.Now().Add(time.Hour)

	meta := model.MediaMeta{
		MediaType:         model.MediaTypeFile,
		HttpContentType:   contentType,
		HttpContentLength: contentLength,
		Name:              name,
		Expiry:            expiry.Unix(),
		UploadURL:         s.beatService.getSaveMediaURL(name, model.MediaTypeFile, expiry),
	}

	s.mediaUploader.On("UploadMedia", ctx, name, contentType, file).Return(nil).Once()

	err := s.beatService.UploadMedia(ctx, file, meta)
	assert.NoError(t, err)
}

func TestUploadMedia_FailURLExpired(t *testing.T) {
	t.Parallel()

	s := createService(t)

	ctx := context.Background()
	contentLength := int64(10)
	expiry := time.Now().Add(-time.Hour)

	meta := model.MediaMeta{
		MediaType:         model.MediaTypeFile,
		HttpContentType:   contentType,
		HttpContentLength: contentLength,
		Name:              name,
		Expiry:            expiry.Unix(),
		UploadURL:         s.beatService.getSaveMediaURL(name, model.MediaTypeFile, expiry),
	}

	err := s.beatService.UploadMedia(ctx, file, meta)
	assert.ErrorIs(t, err, model.ErrURLExpired)
}

func TestUploadMedia_FailSizeExceeded(t *testing.T) {
	t.Parallel()

	s := createService(t)

	ctx := context.Background()
	contentLength := int64(500)
	expiry := time.Now().Add(time.Hour)

	tests := []struct {
		name      string
		mediaType model.MediaType
	}{
		{
			name:      "file",
			mediaType: model.MediaTypeFile,
		},
		{
			name:      "archive",
			mediaType: model.MediaTypeArchive,
		},
		{
			name:      "image",
			mediaType: model.MediaTypeImage,
		},
	}

	meta := model.MediaMeta{
		HttpContentType:   contentType,
		HttpContentLength: contentLength,
		Name:              name,
		Expiry:            expiry.Unix(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta.MediaType = tt.mediaType
			meta.UploadURL = s.beatService.getSaveMediaURL(name, tt.mediaType, expiry)

			err := s.beatService.UploadMedia(ctx, file, meta)
			assert.ErrorIs(t, err, model.ErrSizeExceeded)
		})
	}
}

func TestUploadMedia_FailInvalidHash(t *testing.T) {
	t.Parallel()

	s := createService(t)

	ctx := context.Background()
	contentLength := int64(10)
	expiry := time.Now().Add(time.Hour)

	meta := model.MediaMeta{
		MediaType:         model.MediaTypeFile,
		HttpContentType:   contentType,
		HttpContentLength: contentLength,
		Name:              name,
		Expiry:            expiry.Unix(),
		UploadURL:         "invalid",
	}

	err := s.beatService.UploadMedia(ctx, file, meta)
	assert.ErrorIs(t, err, model.ErrInvalidHash)
}

func TestUpdateBeat_Success(t *testing.T) {
	t.Parallel()

	s := createService(t)

	ctx := context.Background()
	notDownloaded := false
	beat := model.UpdateBeatParams{
		UpdateBeatParams: generated.UpdateBeatParams{
			ID:                uuid.New(),
			IsImageDownloaded: &notDownloaded,
			IsFileDownloaded:  &notDownloaded,
		},
	}

	s.beatModifier.On("UpdateBeat", mock.Anything, beat).Return(&generated.Beat{
		IsImageDownloaded:   false,
		IsFileDownloaded:    false,
		IsArchiveDownloaded: true,
	}, nil).Once()

	file, image, archive, err := s.beatService.UpdateBeat(ctx, beat)
	require.NoError(t, err)
	assert.NotNil(t, file)
	assert.NotNil(t, image)
	assert.Nil(t, archive)
}

func TestUpdateBeat_Fail(t *testing.T) {
	t.Parallel()

	s := createService(t)

	s.beatModifier.On("UpdateBeat", mock.Anything, mock.Anything).Return(nil, model.ErrBeatNotFound).Once()

	_, _, _, err := s.beatService.UpdateBeat(context.Background(), model.UpdateBeatParams{})
	assert.ErrorIs(t, err, model.ErrBeatNotFound)
}

func TestGetBeats_Success(t *testing.T) {
	t.Parallel()

	s := createService(t)

	ctx := context.Background()
	params := model.GetBeatsParams{}
	beats := []model.Beat{
		{
			ID: uuid.New(),
		},
		{
			ID: uuid.New(),
		},
	}
	total := uint64(2)
	url1 := "url1"
	url2 := "url2"

	s.beatProvider.On("GetBeats", mock.Anything, params).Return(beats, &total, nil).Once()
	s.urlProvider.On("GetDownloadMediaURL", mock.Anything, beats[0].ImagePath, time.Minute*time.Duration(s.config.urlTTL)).Return(&url1, nil).Once()
	s.urlProvider.On("GetDownloadMediaURL", mock.Anything, beats[1].ImagePath, time.Minute*time.Duration(s.config.urlTTL)).Return(&url2, nil).Once()

	res, resTotal, err := s.beatService.GetBeats(ctx, params)
	require.NoError(t, err)
	assert.Equal(t, total, *resTotal)
	for i := range beats {
		assert.Equal(t, "url"+strconv.Itoa(i+1), res[i].ImagePath)
		assert.Equal(t, beats[i].ID, res[i].ID)
	}
}

func TestGetBeats_Fail(t *testing.T) {
	t.Parallel()

	s := createService(t)

	expErr := errors.New("error")

	s.beatProvider.On("GetBeats", mock.Anything, mock.Anything).Return(nil, nil, expErr).Once()

	_, _, err := s.beatService.GetBeats(context.Background(), model.GetBeatsParams{})
	assert.ErrorIs(t, expErr, err)
}

func TestGetBeatStream_Success(t *testing.T) {
	t.Parallel()

	s := createService(t)

	ctx := context.Background()
	beatID := uuid.New()
	start := 5
	end := 10
	beat := generated.Beat{
		ID:               beatID,
		FilePath:         uuid.NewString(),
		IsFileDownloaded: true,
	}
	size := 100
	file := io.NopCloser(strings.NewReader("content"))

	s.beatProvider.On("GetBeatByID", mock.Anything, beatID).Return(&beat, nil).Once()
	s.beatBytesProvider.On("GetBeatBytes", mock.Anything, beat.FilePath, &start, &end).Return(file, &size, &contentType, nil).Once()

	resFile, resSlice, resContentType, err := s.beatService.GetBeatStream(ctx, beatID, &start, &end)
	require.NoError(t, err)
	assert.Equal(t, file, resFile)
	if assert.NotNil(t, resSlice) {
		assert.Equal(t, size, *resSlice)
	}
	if assert.NotNil(t, resContentType) {
		assert.Equal(t, contentType, *resContentType)
	}
}

func TestGetBeatStream_FailFileNotDownloaded(t *testing.T) {
	t.Parallel()

	s := createService(t)

	ctx := context.Background()
	beat := generated.Beat{IsFileDownloaded: false}
	start, end := 5, 10

	s.beatProvider.On("GetBeatByID", mock.Anything, mock.Anything).Return(&beat, nil).Once()

	_, _, _, err := s.beatService.GetBeatStream(ctx, uuid.New(), &start, &end)
	assert.ErrorIs(t, err, model.ErrBeatNotFound)
}

func TestGetBeatStream_Fail(t *testing.T) {
	t.Parallel()

	s := createService(t)

	tests := []struct {
		name string
		beh  func()
	}{
		{
			name: "get beat by id error",
			beh: func() {
				s.beatProvider.On("GetBeatByID", mock.Anything, mock.Anything).Return(nil, model.ErrBeatNotFound).Once()
			},
		},
		{
			name: "get beat bytes error",
			beh: func() {
				beat := generated.Beat{IsFileDownloaded: true}
				s.beatProvider.On("GetBeatByID", mock.Anything, mock.Anything).Return(&beat, nil).Once()
				s.beatBytesProvider.On("GetBeatBytes", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil, nil, model.ErrBeatNotFound).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.beh()

			_, _, _, err := s.beatService.GetBeatStream(context.Background(), uuid.New(), nil, nil)
			assert.ErrorIs(t, err, model.ErrBeatNotFound)
		})
	}
}

func TestGetBeatArchive_Success(t *testing.T) {
	t.Parallel()

	s := createService(t)

	ctx := context.Background()
	url := "url"
	params := generated.SaveOwnerParams{
		BeatID: uuid.New(),
		UserID: uuid.New(),
	}
	owner := generated.BeatsOwner(params)
	beat := generated.Beat{
		ID:                  params.BeatID,
		ArchivePath:         uuid.NewString(),
		IsArchiveDownloaded: true,
	}

	s.beatProvider.On("GetBeatByID", mock.Anything, params.BeatID).Return(&beat, nil).Once()
	s.beatProvider.On("GetOwnerByBeatID", mock.Anything, params.BeatID).Return(&owner, nil).Once()
	s.urlProvider.On("GetDownloadMediaURL", mock.Anything, beat.ArchivePath, time.Minute*time.Duration(s.config.urlTTL)).Return(&url, nil).Once()

	res, err := s.beatService.GetBeatArchive(ctx, params)
	require.NoError(t, err)
	assert.Equal(t, url, *res)
}

func TestGetBeatArchive_FailArchiveNotFound(t *testing.T) {
	t.Parallel()

	s := createService(t)
	s.beatProvider.On("GetBeatByID", mock.Anything, mock.Anything).Return(&generated.Beat{}, nil).Once()

	_, err := s.beatService.GetBeatArchive(context.Background(), generated.SaveOwnerParams{})
	assert.ErrorIs(t, err, model.ErrArchiveNotFound)
}

func TestGetBeatArchive_FailInvalidOwner(t *testing.T) {
	t.Parallel()

	s := createService(t)

	ctx := context.Background()
	params := generated.SaveOwnerParams{
		BeatID: uuid.New(),
		UserID: uuid.New(),
	}
	owner := generated.BeatsOwner{
		BeatID: params.BeatID,
		UserID: uuid.New(),
	}

	s.beatProvider.On("GetBeatByID", mock.Anything, params.BeatID).Return(&generated.Beat{IsArchiveDownloaded: true}, nil).Once()
	s.beatProvider.On("GetOwnerByBeatID", mock.Anything, params.BeatID).Return(&owner, nil).Once()

	_, err := s.beatService.GetBeatArchive(ctx, params)
	assert.ErrorIs(t, err, model.ErrInvalidOwner)
}

func TestGetBeatArchive_SuccessOwnerNotFound(t *testing.T) {
	t.Parallel()

	s := createService(t)

	ctx := context.Background()
	params := generated.SaveOwnerParams{
		BeatID: uuid.New(),
		UserID: uuid.New(),
	}

	s.beatProvider.On("GetBeatByID", mock.Anything, params.BeatID).Return(&generated.Beat{IsArchiveDownloaded: true}, nil).Once()
	s.beatProvider.On("GetOwnerByBeatID", mock.Anything, params.BeatID).Return(nil, model.ErrOwnerNotFound).Once()
	s.beatModifier.On("SaveOwner", mock.Anything, params).Return(nil).Once()
	s.urlProvider.On("GetDownloadMediaURL", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil).Once()

	_, err := s.beatService.GetBeatArchive(ctx, params)
	assert.NoError(t, err)
}

func TestGetBeatArchive_Fail(t *testing.T) {
	t.Parallel()

	s := createService(t)

	expErr := errors.New("error")

	tests := []struct {
		name string
		beh  func()
	}{
		{
			name: "get beat by id error",
			beh: func() {
				s.beatProvider.On("GetBeatByID", mock.Anything, mock.Anything).Return(nil, expErr).Once()
			},
		},
		{
			name: "get owner by beat id error",
			beh: func() {
				s.beatProvider.On("GetBeatByID", mock.Anything, mock.Anything).Return(&generated.Beat{IsArchiveDownloaded: true}, nil).Once()
				s.beatProvider.On("GetOwnerByBeatID", mock.Anything, mock.Anything).Return(nil, expErr).Once()
			},
		},
		{
			name: "save owner error",
			beh: func() {
				s.beatProvider.On("GetBeatByID", mock.Anything, mock.Anything).Return(&generated.Beat{IsArchiveDownloaded: true}, nil).Once()
				s.beatProvider.On("GetOwnerByBeatID", mock.Anything, mock.Anything).Return(nil, model.ErrOwnerNotFound).Once()
				s.beatModifier.On("SaveOwner", mock.Anything, mock.Anything).Return(expErr).Once()
			},
		},
		{
			name: "get download media url error",
			beh: func() {
				s.beatProvider.On("GetBeatByID", mock.Anything, mock.Anything).Return(&generated.Beat{IsArchiveDownloaded: true}, nil).Once()
				s.beatProvider.On("GetOwnerByBeatID", mock.Anything, mock.Anything).Return(&generated.BeatsOwner{}, nil).Once()
				s.urlProvider.On("GetDownloadMediaURL", mock.Anything, mock.Anything, mock.Anything).Return(nil, expErr).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.beh()

			_, err := s.beatService.GetBeatArchive(context.Background(), generated.SaveOwnerParams{})
			assert.ErrorIs(t, err, expErr)
		})
	}
}
