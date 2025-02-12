package minio

import (
	"context"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
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

func New(ctx context.Context, config MinioConfig, opts ...Option) (*Minio, error) {
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
		logger.Log().Fatal(ctx, "failed to init minio client: %s", err.Error())
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

		logger.Log().Debug(ctx, "minio failed to check bucket: %s; attempts left: %d", err.Error(), m.connAttempts)

		time.Sleep(m.connTimeout)

		m.connAttempts--
	}
	if err != nil {
		logger.Log().Fatal(ctx, "failed to connect to minio: %s", err.Error())
		return nil, err
	}

	if !exists {
		err = m.Client.MakeBucket(ctx, config.Bucket, minio.MakeBucketOptions{Region: config.Location})
		if err != nil {
			logger.Log().Fatal(ctx, "failed to create bucket: %s", err.Error())
			return nil, err
		}
	}

	return m, nil
}
