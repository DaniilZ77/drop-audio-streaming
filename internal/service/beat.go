package beat

import (
	"context"
	"io"
	"sync"

	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/db/generated"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/model"
	"github.com/google/uuid"
)

type BeatModifier interface {
	SaveBeat(ctx context.Context, beat model.SaveBeatParams) error
	UpdateBeat(ctx context.Context, beat model.UpdateBeatParams) (*generated.Beat, error)
	DeleteBeat(ctx context.Context, id int) error
}

type BeatProvider interface {
	GetBeatByID(ctx context.Context, id int) (*model.Beat, error)
	GetBeats(ctx context.Context, params model.GetBeatsParams) (beats []model.Beat, total int, err error)
	GetBeatParams(ctx context.Context) (params *model.BeatParams, err error)
}

type URLProvider interface {
	GetSaveFileURL(ctx context.Context, path string) (*string, error)
	GetDownloadFileURL(ctx context.Context, path string) (*string, error)
}

type BeatBytesProvider interface {
	GetBeatBytes(ctx context.Context, path string, s, e *int) (file io.ReadCloser, size *int, contentType *string, err error)
}

type BeatService struct {
	beatSaver         BeatModifier
	beatProvider      BeatProvider
	urlProvider       URLProvider
	beatBytesProvider BeatBytesProvider
}

func New(
	beatSaver BeatModifier,
	beatProvider BeatProvider,
	urlProvider URLProvider,
	beatBytesProvider BeatBytesProvider,
) *BeatService {
	return &BeatService{
		beatSaver:         beatSaver,
		beatProvider:      beatProvider,
		urlProvider:       urlProvider,
		beatBytesProvider: beatBytesProvider,
	}
}

func (s *BeatService) GetBeatStream(ctx context.Context, beatID int, start, end *int) (file io.ReadCloser, size *int, contentType *string, err error) {
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

func (s *BeatService) StreamBeat(ctx context.Context, r io.Reader, w io.Writer, chunkSize int) error {
	data := make(chan []byte)
	quit := make(chan struct{}, 1)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer close(data)
		defer wg.Done()

		for {
			buf := make([]byte, chunkSize)
			n, err := r.Read(buf)
			if err != nil && err != io.EOF {
				logger.Log().Error(ctx, err.Error())
				return
			}

			if n == 0 {
				return
			}

			select {
			case data <- buf[:n]:
			case <-quit:
				return
			}
		}
	}()

	go func() {
		defer func() { quit <- struct{}{} }()
		defer wg.Done()
		for chunk := range data {
			if _, err := w.Write(chunk); err != nil {
				logger.Log().Error(ctx, err.Error())
				return
			}
		}
	}()

	wg.Wait()

	return nil
}

func (s *BeatService) SaveBeat(ctx context.Context, beat model.SaveBeatParams) (fileUploadURL, imageUploadURL *string, err error) {
	filePath := uuid.New().String()
	beat.FilePath = filePath

	imagePath := uuid.New().String()
	beat.ImagePath = imagePath

	err = s.beatSaver.SaveBeat(ctx, beat)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	fileUploadURL, err = s.urlProvider.GetSaveFileURL(ctx, filePath)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	imageUploadURL, err = s.urlProvider.GetSaveFileURL(ctx, imagePath)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	return
}

func (s *BeatService) GetBeats(ctx context.Context, params model.GetBeatsParams) (beats []model.Beat, total int, err error) {
	beats, total, err = s.beatProvider.GetBeats(ctx, params)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, 0, err
	}

	for i, v := range beats {
		url, err := s.urlProvider.GetDownloadFileURL(ctx, v.ImagePath)
		if err != nil {
			logger.Log().Error(ctx, err.Error())
			return nil, 0, err
		}

		beats[i].ImagePath = *url
	}

	return beats, total, nil
}

func (s *BeatService) GetBeatParams(ctx context.Context) (params *model.BeatParams, err error) {
	return s.beatProvider.GetBeatParams(ctx)
}

func (s *BeatService) UpdateBeat(ctx context.Context, updateBeat model.UpdateBeatParams) (fileUploadURL, imageUploadURL *string, err error) {
	beat, err := s.beatSaver.UpdateBeat(ctx, updateBeat)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	if !beat.IsFileDownloaded {
		fileUploadURL, err = s.urlProvider.GetSaveFileURL(ctx, beat.FilePath)
		if err != nil {
			logger.Log().Error(ctx, err.Error())
			return nil, nil, err
		}
	}

	if !beat.IsImageDownloaded {
		imageUploadURL, err = s.urlProvider.GetSaveFileURL(ctx, beat.ImagePath)
		if err != nil {
			logger.Log().Error(ctx, err.Error())
			return nil, nil, err
		}
	}

	return fileUploadURL, imageUploadURL, nil
}

func (s *BeatService) DeleteBeat(ctx context.Context, id int) error {
	return s.beatSaver.DeleteBeat(ctx, id)
}
