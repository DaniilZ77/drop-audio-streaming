package beat

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/db/generated"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/domain/model"
	sl "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
	"github.com/google/uuid"
)

type BeatServiceConfig struct {
	fileSizeLimit      int64
	archiveSizeLimit   int64
	imageSizeLimit     int64
	verificationSecret string
	urlTTL             int
}

func NewBeatServiceConfig(fileSizeLimit int64, archiveSizeLimit int64, imageSizeLimit int64, verificationSecret string, urlTTL int) *BeatServiceConfig {
	return &BeatServiceConfig{
		fileSizeLimit:      fileSizeLimit,
		archiveSizeLimit:   archiveSizeLimit,
		imageSizeLimit:     imageSizeLimit,
		verificationSecret: verificationSecret,
		urlTTL:             urlTTL,
	}
}

//go:generate mockery --name BeatModifier
type BeatModifier interface {
	SaveBeat(ctx context.Context, beat model.SaveBeat) error
	UpdateBeat(ctx context.Context, beat model.UpdateBeat) (*generated.Beat, error)
	DeleteBeat(ctx context.Context, id uuid.UUID) error
	SaveOwner(ctx context.Context, owner generated.SaveOwnerParams) error
}

//go:generate mockery --name BeatProvider
type BeatProvider interface {
	GetBeatByID(ctx context.Context, id uuid.UUID) (*generated.Beat, error)
	GetBeats(ctx context.Context, params model.GetBeatsParams) (beats []model.Beat, total *uint64, err error)
	GetBeatParams(ctx context.Context) (attrs *model.BeatAttributes, err error)
	GetOwnerByBeatID(ctx context.Context, beatID uuid.UUID) (*generated.BeatsOwner, error)
}

//go:generate mockery --name URLProvider
type URLProvider interface {
	GetDownloadMediaURL(ctx context.Context, path string, expires time.Duration) (*string, error)
}

//go:generate mockery --name BeatBytesProvider
type BeatBytesProvider interface {
	GetBeatBytes(ctx context.Context, path string, s, e *int) (file io.ReadCloser, size *int, contentType *string, err error)
}

//go:generate mockery --name MediaUploader
type MediaUploader interface {
	UploadMedia(ctx context.Context, path, contentType string, file io.Reader) error
}

type BeatService struct {
	beatModifier      BeatModifier
	beatProvider      BeatProvider
	urlProvider       URLProvider
	mediaUploader     MediaUploader
	beatBytesProvider BeatBytesProvider
	config            *BeatServiceConfig
	log               *slog.Logger
}

func NewBeatService(
	beatSaver BeatModifier,
	beatProvider BeatProvider,
	urlProvider URLProvider,
	mediaUploader MediaUploader,
	beatBytesProvider BeatBytesProvider,
	config *BeatServiceConfig,
	log *slog.Logger,
) *BeatService {
	return &BeatService{
		beatModifier:      beatSaver,
		beatProvider:      beatProvider,
		urlProvider:       urlProvider,
		mediaUploader:     mediaUploader,
		beatBytesProvider: beatBytesProvider,
		config:            config,
		log:               log,
	}
}

func (s *BeatService) GetBeatStream(ctx context.Context, beatID uuid.UUID, start, end *int) (file io.ReadCloser, size *int, contentType *string, err error) {
	beat, err := s.beatProvider.GetBeatByID(ctx, beatID)
	if err != nil {
		s.log.Error("failed to get beat", sl.Err(err))
		return nil, nil, nil, err
	}

	if !beat.IsFileDownloaded {
		s.log.Debug("file is not downloaded")
		return nil, nil, nil, &model.ModelError{Err: model.ErrBeatNotFound}
	}

	file, size, contentType, err = s.beatBytesProvider.GetBeatBytes(ctx, beat.FilePath, start, end)
	if err != nil {
		s.log.Error("failed to get beat bytes", sl.Err(err))
		return nil, nil, nil, err
	}

	return file, size, contentType, nil
}

func (s *BeatService) getSaveMediaURL(name string, mt model.MediaType, exp time.Time) string {
	url := "/v1/beat?"

	url += fmt.Sprintf("name=%s", name)
	url += fmt.Sprintf("&type=%s", mt)
	url += fmt.Sprintf("&exp=%d", exp.Unix())

	mac := hmac.New(sha1.New, []byte(s.config.verificationSecret))
	mac.Write([]byte(url))

	sig := base64.URLEncoding.EncodeToString(mac.Sum(nil))
	url += fmt.Sprintf("&hash=%s", sig)

	return url
}

func (s *BeatService) SaveBeat(ctx context.Context, beat model.SaveBeat) (*string, *string, *string, error) {
	filePath := uuid.New().String()
	beat.FilePath = filePath

	imagePath := uuid.New().String()
	beat.ImagePath = imagePath

	archivePath := uuid.New().String()
	beat.ArchivePath = archivePath

	err := s.beatModifier.SaveBeat(ctx, beat)
	if err != nil {
		s.log.Error("failed to save beat", sl.Err(err))
		return nil, nil, nil, err
	}

	exp := time.Now().Add(time.Minute * time.Duration(s.config.urlTTL))
	fileUploadURL := s.getSaveMediaURL(filePath, model.MediaTypeFile, exp)
	imageUploadURL := s.getSaveMediaURL(imagePath, model.MediaTypeImage, exp)
	archiveUploadURL := s.getSaveMediaURL(archivePath, model.MediaTypeArchive, exp)

	return &fileUploadURL, &imageUploadURL, &archiveUploadURL, nil
}

func (s *BeatService) GetBeats(ctx context.Context, params model.GetBeatsParams) (beats []model.Beat, total *uint64, err error) {
	beats, total, err = s.beatProvider.GetBeats(ctx, params)
	if err != nil {
		s.log.Error("failed to get beats", sl.Err(err))
		return nil, nil, err
	}

	for i, v := range beats {
		url, err := s.urlProvider.GetDownloadMediaURL(ctx, v.ImagePath, time.Minute*time.Duration(s.config.urlTTL))
		if err != nil {
			s.log.Error("failed to get download media url", sl.Err(err))
			return nil, nil, err
		}

		beats[i].ImagePath = *url
	}

	return beats, total, nil
}

func (s *BeatService) GetBeatParams(ctx context.Context) (attrs *model.BeatAttributes, err error) {
	return s.beatProvider.GetBeatParams(ctx)
}

func (s *BeatService) UpdateBeat(ctx context.Context, updateBeat model.UpdateBeat) (*string, *string, *string, error) {
	beat, err := s.beatModifier.UpdateBeat(ctx, updateBeat)
	if err != nil {
		s.log.Error("failed to update beat", sl.Err(err))
		return nil, nil, nil, err
	}

	exp := time.Now().Add(time.Minute * time.Duration(s.config.urlTTL))
	var fileUploadURL, imageUploadURL, archiveUploadURL *string
	if !beat.IsFileDownloaded {
		fileUploadURL = new(string)
		*fileUploadURL = s.getSaveMediaURL(beat.FilePath, model.MediaTypeFile, exp)
	}

	if !beat.IsImageDownloaded {
		imageUploadURL = new(string)
		*imageUploadURL = s.getSaveMediaURL(beat.ImagePath, model.MediaTypeImage, exp)
	}

	if !beat.IsArchiveDownloaded {
		archiveUploadURL = new(string)
		*archiveUploadURL = s.getSaveMediaURL(beat.ArchivePath, model.MediaTypeArchive, exp)
	}

	return fileUploadURL, imageUploadURL, archiveUploadURL, nil
}

func (s *BeatService) DeleteBeat(ctx context.Context, id uuid.UUID) error {
	return s.beatModifier.DeleteBeat(ctx, id)
}

func (s *BeatService) UploadMedia(ctx context.Context, file io.Reader, m model.MediaMeta) error {
	if m.Expiry < time.Now().Unix() {
		s.log.Debug("url expired", slog.Int64("expiry", m.Expiry), slog.Int64("now", time.Now().Unix()))
		return &model.ModelError{Err: model.ErrURLExpired}
	}

	switch m.MediaType {
	case model.MediaTypeFile:
		if m.HttpContentLength > s.config.fileSizeLimit {
			s.log.Debug("file size exceeded", slog.Int64("size", m.HttpContentLength), slog.Int64("limit", s.config.fileSizeLimit))
			return model.NewErr(model.ErrSizeExceeded, fmt.Sprintf("file, %d > %d", m.HttpContentLength, s.config.fileSizeLimit))
		}
	case model.MediaTypeArchive:
		if m.HttpContentLength > s.config.archiveSizeLimit {
			s.log.Debug("archive size exceeded", slog.Int64("size", m.HttpContentLength), slog.Int64("limit", s.config.archiveSizeLimit))
			return model.NewErr(model.ErrSizeExceeded, fmt.Sprintf("archive, %d > %d", m.HttpContentLength, s.config.archiveSizeLimit))
		}
	case model.MediaTypeImage:
		if m.HttpContentLength > s.config.imageSizeLimit {
			s.log.Debug("image size exceeded", slog.Int64("size", m.HttpContentLength), slog.Int64("limit", s.config.imageSizeLimit))
			return model.NewErr(model.ErrSizeExceeded, fmt.Sprintf("image, %d > %d", m.HttpContentLength, s.config.imageSizeLimit))
		}
	}

	url := s.getSaveMediaURL(m.Name, m.MediaType, time.Unix(m.Expiry, 0))
	if url != m.UploadURL {
		s.log.Debug("invalid url", slog.String("url", m.UploadURL), slog.String("expected", url))
		return &model.ModelError{Err: model.ErrInvalidHash}
	}

	if err := s.mediaUploader.UploadMedia(ctx, m.Name, m.HttpContentType, file); err != nil {
		s.log.Error("failed to upload media", sl.Err(err))
		return err
	}

	return nil
}

func (s *BeatService) GetBeatArchive(ctx context.Context, params generated.SaveOwnerParams) (*string, error) {
	beat, err := s.beatProvider.GetBeatByID(ctx, params.BeatID)
	if err != nil {
		s.log.Error("failed to get beat", sl.Err(err))
		return nil, err
	}

	if !beat.IsArchiveDownloaded {
		s.log.Debug("archive not found")
		return nil, &model.ModelError{Err: model.ErrArchiveNotFound}
	}

	owner, err := s.beatProvider.GetOwnerByBeatID(ctx, params.BeatID)
	if err != nil && !errors.Is(err, model.ErrOwnerNotFound) {
		s.log.Error("failed to get owner", sl.Err(err))
		return nil, err
	}

	if errors.Is(err, model.ErrOwnerNotFound) {
		if err := s.beatModifier.SaveOwner(ctx, params); err != nil {
			s.log.Error("failed to save owner", sl.Err(err))
			return nil, err
		}
	} else if params.UserID != owner.UserID {
		s.log.Debug("beat acquired by another owner", slog.String("beat_id", params.BeatID.String()), slog.String("user_id", params.UserID.String()), slog.String("owner_id", owner.UserID.String()))
		return nil, model.NewErr(model.ErrInvalidOwner, "beat acquired by another owner")
	}

	url, err := s.urlProvider.GetDownloadMediaURL(ctx, beat.ArchivePath, time.Minute*time.Duration(s.config.urlTTL))
	if err != nil {
		s.log.Error("failed to get download media url", sl.Err(err))
		return nil, err
	}

	return url, nil
}
