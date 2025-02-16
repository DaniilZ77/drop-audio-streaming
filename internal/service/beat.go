package beat

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/db/generated"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/model"
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

type BeatModifier interface {
	SaveBeat(ctx context.Context, beat model.SaveBeatParams) error
	UpdateBeat(ctx context.Context, beat model.UpdateBeatParams) (*generated.Beat, error)
	DeleteBeat(ctx context.Context, id uuid.UUID) error
	SaveOwner(ctx context.Context, owner generated.SaveOwnerParams) error
}

type BeatProvider interface {
	GetBeatByID(ctx context.Context, id uuid.UUID) (*generated.Beat, error)
	GetBeats(ctx context.Context, params model.GetBeatsParams) (beats []model.Beat, total *int, err error)
	GetBeatParams(ctx context.Context) (params *model.BeatParams, err error)
	GetOwnerByBeatID(ctx context.Context, beatID uuid.UUID) (*generated.BeatsOwner, error)
}

type URLProvider interface {
	GetDownloadMediaURL(ctx context.Context, path string, expires time.Duration) (*string, error)
}

type BeatBytesProvider interface {
	GetBeatBytes(ctx context.Context, path string, s, e *int) (file io.ReadCloser, size *int, contentType *string, err error)
}

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
}

func NewBeatService(
	beatSaver BeatModifier,
	beatProvider BeatProvider,
	urlProvider URLProvider,
	mediaUploader MediaUploader,
	beatBytesProvider BeatBytesProvider,
	config *BeatServiceConfig,
) *BeatService {
	return &BeatService{
		beatModifier:      beatSaver,
		beatProvider:      beatProvider,
		urlProvider:       urlProvider,
		mediaUploader:     mediaUploader,
		beatBytesProvider: beatBytesProvider,
		config:            config,
	}
}

func (s *BeatService) GetBeatStream(ctx context.Context, beatID uuid.UUID, start, end *int) (file io.ReadCloser, size *int, contentType *string, err error) {
	beat, err := s.beatProvider.GetBeatByID(ctx, beatID)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, nil, err
	}

	file, size, contentType, err = s.beatBytesProvider.GetBeatBytes(ctx, beat.FilePath, start, end)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
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

func (s *BeatService) SaveBeat(ctx context.Context, beat model.SaveBeatParams) (*string, *string, *string, error) {
	filePath := uuid.New().String()
	beat.FilePath = filePath

	imagePath := uuid.New().String()
	beat.ImagePath = imagePath

	archivePath := uuid.New().String()
	beat.ArchivePath = archivePath

	err := s.beatModifier.SaveBeat(ctx, beat)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, nil, err
	}

	exp := time.Now().Add(time.Minute * time.Duration(s.config.urlTTL))
	fileUploadURL := s.getSaveMediaURL(filePath, model.File, exp)
	imageUploadURL := s.getSaveMediaURL(imagePath, model.Image, exp)
	archiveUploadURL := s.getSaveMediaURL(archivePath, model.Archive, exp)

	return &fileUploadURL, &imageUploadURL, &archiveUploadURL, nil
}

func (s *BeatService) GetBeats(ctx context.Context, params model.GetBeatsParams) (beats []model.Beat, total *int, err error) {
	beats, total, err = s.beatProvider.GetBeats(ctx, params)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	for i, v := range beats {
		url, err := s.urlProvider.GetDownloadMediaURL(ctx, v.ImagePath, time.Minute*time.Duration(s.config.urlTTL))
		if err != nil {
			logger.Log().Error(ctx, err.Error())
			return nil, nil, err
		}

		beats[i].ImagePath = *url
	}

	return beats, total, nil
}

func (s *BeatService) GetBeatParams(ctx context.Context) (params *model.BeatParams, err error) {
	return s.beatProvider.GetBeatParams(ctx)
}

func (s *BeatService) UpdateBeat(ctx context.Context, updateBeat model.UpdateBeatParams) (*string, *string, *string, error) {
	beat, err := s.beatModifier.UpdateBeat(ctx, updateBeat)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, nil, err
	}

	exp := time.Now().Add(time.Minute * time.Duration(s.config.urlTTL))
	var fileUploadURL, imageUploadURL, archiveUploadURL string
	if !beat.IsFileDownloaded {
		fileUploadURL = s.getSaveMediaURL(beat.FilePath, model.File, exp)
	}

	if !beat.IsImageDownloaded {
		imageUploadURL = s.getSaveMediaURL(beat.ImagePath, model.Image, exp)
	}

	if !beat.IsArchiveDownloaded {
		archiveUploadURL = s.getSaveMediaURL(beat.ArchivePath, model.Archive, exp)
	}

	return &fileUploadURL, &imageUploadURL, &archiveUploadURL, nil
}

func (s *BeatService) DeleteBeat(ctx context.Context, id uuid.UUID) error {
	return s.beatModifier.DeleteBeat(ctx, id)
}

func (s *BeatService) UploadMedia(ctx context.Context, file io.Reader, m model.MediaMeta) error {
	if m.Expiry < time.Now().Unix() {
		logger.Log().Error(ctx, model.ErrURLExpired.Error())
		return &model.ModelError{Err: model.ErrURLExpired}
	}

	switch m.MediaType {
	case model.File:
		if m.ContentLength > s.config.fileSizeLimit {
			logger.Log().Debug(ctx, model.ErrFileSizeExceeded.Error())
			return model.NewErr(model.ErrFileSizeExceeded, fmt.Sprintf("%d > %d", m.ContentLength, s.config.fileSizeLimit))
		}
	case model.Archive:
		if m.ContentLength > s.config.archiveSizeLimit {
			logger.Log().Debug(ctx, model.ErrArchiveSizeExceeded.Error())
			return model.NewErr(model.ErrArchiveSizeExceeded, fmt.Sprintf("%d > %d", m.ContentLength, s.config.archiveSizeLimit))
		}
	case model.Image:
		if m.ContentLength > s.config.imageSizeLimit {
			logger.Log().Debug(ctx, model.ErrImageSizeExceeded.Error())
			return model.NewErr(model.ErrImageSizeExceeded, fmt.Sprintf("%d > %d", m.ContentLength, s.config.imageSizeLimit))
		}
	}

	url := s.getSaveMediaURL(m.Name, m.MediaType, time.Unix(m.Expiry, 0))
	if url != m.URL {
		logger.Log().Debug(ctx, model.ErrInvalidHash.Error())
		return &model.ModelError{Err: model.ErrInvalidHash}
	}

	if err := s.mediaUploader.UploadMedia(ctx, m.Name, m.ContentType, file); err != nil {
		logger.Log().Error(ctx, err.Error())
		return err
	}

	return nil
}

func (s *BeatService) GetBeatArchive(ctx context.Context, params generated.SaveOwnerParams) (*string, error) {
	beat, err := s.beatProvider.GetBeatByID(ctx, params.BeatID)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	if !beat.IsArchiveDownloaded {
		logger.Log().Debug(ctx, model.ErrArchiveNotFound.Error())
		return nil, &model.ModelError{Err: model.ErrArchiveNotFound}
	}

	owner, err := s.beatProvider.GetOwnerByBeatID(ctx, params.BeatID)
	if err != nil && !errors.Is(err, model.ErrOwnerNotFound) {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	if errors.Is(err, model.ErrOwnerNotFound) {
		if err := s.beatModifier.SaveOwner(ctx, params); err != nil {
			logger.Log().Error(ctx, err.Error())
			return nil, err
		}
	} else if params.UserID != owner.UserID {
		logger.Log().Debug(ctx, model.ErrInvalidOwner.Error())
		return nil, model.NewErr(model.ErrInvalidOwner, "beat acquired by another owner")
	}

	url, err := s.urlProvider.GetDownloadMediaURL(ctx, beat.ArchivePath, time.Minute*time.Duration(s.config.urlTTL))
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	return url, nil
}
