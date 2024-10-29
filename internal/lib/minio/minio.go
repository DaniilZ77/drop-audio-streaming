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
	// TODO: figure out how to set max pool size

	m := &Minio{
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
	}

	// Custom options
	for _, opt := range opts {
		opt(m)
	}

	var client *minio.Client
	var err error

	for m.connAttempts > 0 {
		client, err = minio.New(config.Endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(config.User, config.Password, ""),
			Secure: config.UseSSL,
		})
		if err == nil {
			m.Client = client
			break
		}

		logger.Log().Debug(ctx, "failed to connect to minio: %s", err.Error())

		time.Sleep(m.connTimeout)

		m.connAttempts--
	}

	if err != nil {
		logger.Log().Fatal(ctx, "failed to connect to minio: %s", err.Error())
		return nil, err
	}

	exists, err := client.BucketExists(ctx, config.Bucket)
	if err != nil {
		logger.Log().Fatal(ctx, "failed to check bucket: %s", err.Error())
		return nil, err
	}

	if !exists {
		err = client.MakeBucket(ctx, config.Bucket, minio.MakeBucketOptions{Region: config.Location})
		if err != nil {
			logger.Log().Fatal(ctx, "failed to create bucket: %s", err.Error())
			return nil, err
		}
	}

	return m, nil
}
