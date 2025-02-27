package app

import (
	"context"
	"log/slog"

	grpcapp "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/app/grpc"
	httpapp "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/app/http"
	client "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/client"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/config"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/minio"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/postgres"
	beat "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/service"
	beatstore "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/store"
)

type App struct {
	GRPCServer *grpcapp.App
	Pg         *postgres.Postgres
	HTTPServer *httpapp.App
	Mio        *minio.Minio
}

func New(ctx context.Context,
	cfg *config.Config,
	log *slog.Logger) *App {
	// Postgres connection
	pg, err := postgres.New(ctx, cfg.DatabaseURL, log)
	if err != nil {
		panic(err)
	}

	// Minio connection
	mio, err := minio.New(ctx, minio.MinioConfig{
		Password: cfg.Minio.Password,
		User:     cfg.Minio.User,
		Endpoint: cfg.Minio.Url,
		Bucket:   cfg.Minio.Bucket,
	}, log)
	if err != nil {
		panic(err)
	}

	// Store
	beatStore := beatstore.New(mio, pg, cfg.Minio.Bucket, log)

	// Service
	beatServiceConfig := beat.NewBeatServiceConfig(
		cfg.FileSizeLimit,
		cfg.ArchiveSizeLimit,
		cfg.ImageSizeLimit,
		cfg.VerificationSecret,
		cfg.UrlTtl)
	beatService := beat.NewBeatService(
		beatStore,
		beatStore,
		beatStore,
		beatStore,
		beatStore,
		beatServiceConfig,
		log)

	// gRPC client
	gRPCUserClient, err := client.NewUserClient(ctx,
		cfg.GrpcClient.Port,
		cfg.GrpcClient.Timeout,
		cfg.GrpcClient.Retries,
		log)
	if err != nil {
		panic(err)
	}

	if err := gRPCUserClient.Health(ctx); err != nil {
		panic(err)
	}

	// gRPC server
	gRPCApp := grpcapp.New(ctx, cfg, beatService, gRPCUserClient, log)

	// HTTP server
	httpApp := httpapp.New(ctx, cfg, beatService, gRPCUserClient, log)

	return &App{
		GRPCServer: gRPCApp,
		Pg:         pg,
		Mio:        mio,
		HTTPServer: httpApp,
	}
}
