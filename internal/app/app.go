package app

import (
	"context"

	grpcapp "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/app/grpc"
	httpapp "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/app/http"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/config"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/minio"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/postgres"
	beat "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/service"
	beatstore "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/store/postgres"
)

type App struct {
	GRPCServer *grpcapp.App
	PG         *postgres.Postgres
	HTTPServer *httpapp.App
	M          *minio.Minio
}

func New(ctx context.Context, cfg *config.Config) *App {
	// Init logger
	logger.New(cfg.Log.Level)

	// Postgres connection
	pg, err := postgres.New(ctx, cfg.DB.URL)
	if err != nil {
		logger.Log().Fatal(ctx, "error with connection to database: %s", err.Error())
	}

	// Minio connection
	minio, err := minio.New(ctx, minio.MinioConfig{
		Password: cfg.DB.MinioPassword,
		User:     cfg.DB.MinioUser,
		Endpoint: cfg.DB.MinioEndpoint,
		Bucket:   cfg.DB.MinioBucket,
	})
	if err != nil {
		logger.Log().Fatal(ctx, "error with connection to minio: %s", err.Error())
	}

	// Store
	beatStore := beatstore.New(minio, pg, cfg.DB.MinioBucket)

	// Service
	beatService := beat.New(beatStore)

	// gRPC server
	gRPCApp := grpcapp.New(ctx, cfg, beatService)

	// HTTP server
	httpApp := httpapp.New(ctx, cfg, beatService)

	return &App{
		GRPCServer: gRPCApp,
		PG:         pg,
		M:          minio,
		HTTPServer: httpApp,
	}
}
