package app

import (
	"context"

	grpcapp "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/app/grpc"
	httpapp "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/app/http"
	client "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/client"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/config"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
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

func New(ctx context.Context, cfg *config.Config) *App {
	// Init logger
	logger.New(cfg.Log.Level)

	// Postgres connection
	pg, err := postgres.New(ctx, cfg.DB.URL)
	if err != nil {
		logger.Log().Fatal(ctx, "error with connection to database: %s", err.Error())
		return nil
	}

	// Minio connection
	mio, err := minio.New(ctx, minio.MinioConfig{
		Password: cfg.DB.MinioPassword,
		User:     cfg.DB.MinioUser,
		Endpoint: cfg.DB.MinioEndpoint,
		Bucket:   cfg.DB.MinioBucket,
	})
	if err != nil {
		logger.Log().Fatal(ctx, "error with connection to minio: %s", err.Error())
		return nil
	}

	// Store
	beatStore := beatstore.New(mio, pg, cfg.DB.MinioBucket)

	// Service
	beatServiceConfig := beat.NewBeatServiceConfig(cfg.FileSizeLimit, cfg.ArchiveSizeLimit, cfg.ImageSizeLimit, cfg.VerificationSecret, cfg.URLTTL)
	beatService := beat.NewBeatService(beatStore, beatStore, beatStore, beatStore, beatStore, beatServiceConfig)

	// gRPC client
	gRPCUserClient, err := client.NewUserClient(
		ctx,
		cfg.GRPCUserClientAddr,
		cfg.GRPCClientTimeout,
		cfg.GRPCClientRetries,
	)
	if err != nil {
		logger.Log().Fatal(ctx, "error with connection to user grpc server: %s", err.Error())
		return nil
	}

	// gRPC server
	gRPCApp := grpcapp.New(ctx, cfg, beatService, gRPCUserClient)

	// HTTP server
	httpApp := httpapp.New(ctx, cfg, beatService, gRPCUserClient)

	return &App{
		GRPCServer: gRPCApp,
		Pg:         pg,
		Mio:        mio,
		HTTPServer: httpApp,
	}
}
