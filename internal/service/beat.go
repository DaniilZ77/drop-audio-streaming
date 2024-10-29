package beat

import (
	"context"
	"io"
	"sync"

	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
	minio "github.com/minio/minio-go/v7"
)

type service struct {
	beatStore core.BeatStorage
}

func New(beatStore core.BeatStorage) core.BeatService {
	return &service{beatStore: beatStore}
}

func (s *service) GetBeat(ctx context.Context, beatID int, start, end int64) (*minio.Object, int, error) {
	beat, err := s.beatStore.GetBeatByID(ctx, beatID)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, 0, err
	}

	obj, size, err := s.beatStore.GetBeatFromS3(ctx, beat.Path, start, end)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, 0, err
	}

	logger.Log().Debug(ctx, "size of file: %d", size)

	return obj, size, err
}

func (s *service) WritePartialContent(ctx context.Context, r io.Reader, w io.Writer, chunkSize int) error {
	data := make(chan []byte)
	var wg sync.WaitGroup
	wg.Add(1)

	defer close(data)
	defer wg.Wait()

	go func() {
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

			data <- buf[:n]
		}
	}()

	go func() {
		for chunk := range data {
			if _, err := w.Write(chunk); err != nil {
				return
			}
		}
	}()

	return nil
}
