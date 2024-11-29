package beat

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
	"github.com/google/uuid"
)

type service struct {
	beatStore    core.BeatStorage
	uploadURLTTL int
}

func New(beatStore core.BeatStorage, uploadURLTTL int) core.BeatService {
	return &service{beatStore: beatStore, uploadURLTTL: uploadURLTTL}
}

func (s *service) GetBeatFromS3(ctx context.Context, beatID, start int, end *int) (obj io.ReadCloser, size int, contentType string, err error) {
	beat, err := s.beatStore.GetBeatByID(ctx, beatID, core.True)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, 0, "", err
	}

	obj, size, contentType, err = s.beatStore.GetBeatFromS3(ctx, beat.FilePath, start, end)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, 0, "", err
	}

	logger.Log().Debug(ctx, "size of file: %d", size)

	return obj, size, contentType, nil
}

func (s *service) WritePartialContent(ctx context.Context, r io.Reader, w io.Writer, chunkSize int) error {
	data := make(chan []byte)
	quit := make(chan struct{}, 1)
	var wg sync.WaitGroup
	wg.Add(2)
	logger.Log().Debug(ctx, "beat stream started")

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
	logger.Log().Debug(ctx, "beat stream ended")

	return nil
}

func (s *service) AddBeat(ctx context.Context, beat core.Beat, beatGenre []core.BeatGenre) (filePath, imagePath string, err error) {
	filePath = uuid.New().String()
	beat.FilePath = filePath

	imagePath = uuid.New().String()
	beat.ImagePath = imagePath

	_, err = s.beatStore.AddBeat(ctx, beat, beatGenre)
	if err != nil {
		if errors.Is(err, core.ErrBeatExists) {
			beat, err := s.beatStore.GetBeatByID(ctx, beat.ID, core.Any)
			if err != nil {
				return "", "", err
			}

			return beat.FilePath, beat.ImagePath, nil
		}
		return "", "", err
	}

	return filePath, imagePath, nil
}

func (s *service) GetUploadURL(ctx context.Context, beatPath string) (string, error) {
	path, err := s.beatStore.GetPresignedURL(ctx, beatPath, time.Duration(s.uploadURLTTL)*time.Minute)
	if err != nil {
		return "", err
	}

	return path, nil
}

func (s *service) GetBeatByFilter(ctx context.Context, userID int, filter core.FeedFilter) (beat *core.Beat, genre *string, err error) {
	beats, err := s.beatStore.GetUserSeenBeats(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	beat, genre, err = s.beatStore.GetBeatByFilter(ctx, filter, beats)
	if err != nil {
		if errors.Is(err, core.ErrBeatNotFound) {
			if err := s.beatStore.ClearUserSeenBeats(ctx, userID); err != nil {
				return nil, nil, err
			}

			beat, genre, err = s.beatStore.GetBeatByFilter(ctx, filter, []string{})
			if err != nil {
				return nil, nil, err
			}
		} else {
			return nil, nil, err
		}
	}

	if err := s.beatStore.ReplaceUserSeenBeat(ctx, userID, beat.ID); err != nil {
		return nil, nil, err
	}

	return beat, genre, nil
}

func (s *service) GetBeat(ctx context.Context, beatID int) (beat *core.Beat, beatGenres []core.BeatGenre, err error) {
	beat, err = s.beatStore.GetBeatByID(ctx, beatID, core.True)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	beatGenres, err = s.beatStore.GetBeatGenres(ctx, beatID)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	return beat, beatGenres, nil
}

func (s *service) GetBeatsByBeatmakerID(ctx context.Context, beatmakerID int, p core.GetBeatsParams) (beats []core.Beat, beatsGenres [][]core.BeatGenre, total int, err error) {
	beats, total, err = s.beatStore.GetBeatsByBeatmakerID(ctx, beatmakerID, p)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, 0, err
	}

	for _, beat := range beats {
		genres, err := s.beatStore.GetBeatGenres(ctx, beat.ID)
		if err != nil {
			logger.Log().Error(ctx, err.Error())
			return nil, nil, 0, err
		}
		beatsGenres = append(beatsGenres, genres)
	}

	return beats, beatsGenres, total, nil
}
