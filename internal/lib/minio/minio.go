package minio

import (
	"context"
	"log/slog"
	"time"

	sl "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second
)

type MinioConfig struct {
	Password string
	User     string
	Endpoint string
	Bucket   string
	UseSSL   bool
	Location string
}

type Minio struct {
	connAttempts int
	connTimeout  time.Duration

	Client *minio.Client
}

func New(ctx context.Context, config MinioConfig, log *slog.Logger, opts ...Option) (*Minio, error) {
	m := &Minio{
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
	}

	// Custom options
	for _, opt := range opts {
		opt(m)
	}

	var err error
	m.Client, err = minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.User, config.Password, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		log.Error("failed to init minio client", sl.Err(err))
		return nil, err
	}

	var exists bool

	for m.connAttempts > 0 {
		ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		exists, err = m.Client.BucketExists(ctx, config.Bucket)
		cancel()
		if err == nil {
			break
		}

		log.Debug("minio failed to check bucket", slog.Int("attempts left", m.connAttempts), sl.Err(err))

		time.Sleep(m.connTimeout)

		m.connAttempts--
	}
	if err != nil {
		log.Error("failed to connect to minio", sl.Err(err))
		return nil, err
	}

	if !exists {
		err = m.Client.MakeBucket(ctx, config.Bucket, minio.MakeBucketOptions{Region: config.Location})
		if err != nil {
			log.Error("failed to create bucket", sl.Err(err))
			return nil, err
		}
	}

	return m, nil
}
