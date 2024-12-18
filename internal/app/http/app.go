package http

import (
	"context"
	"net/http"
	"time"

	audiov1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/audio"
	userclient "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/client/user/grpc"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/config"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core"
	router "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/http"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	httpServer *http.Server
	cert       string
	key        string
}

func New(
	ctx context.Context,
	cfg *config.Config,
	beatService core.BeatService,
	grpcUserClient *userclient.Client,
) *App {
	// creds, err := credentials.NewClientTLSFromFile(cfg.Cert, "") nolint
	// if err != nil {
	// 	logger.Log().Fatal(ctx, "failed to create server TLS credentials: %v", err)
	// }

	conn, err := grpc.NewClient(cfg.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log().Fatal(ctx, "failed to dial server:", err)
	}

	gwmux := runtime.NewServeMux()
	router.NewRouter(gwmux, beatService, cfg.ChunkSize, cfg.JWTSecret, grpcUserClient)

	// Register user
	err = audiov1.RegisterAudioServiceHandler(ctx, gwmux, conn)
	if err != nil {
		logger.Log().Fatal(ctx, "failed to register gateway: %w", err)
	}

	// Cors
	withCors := cors.AllowAll().Handler(gwmux)

	// Server
	gwServer := &http.Server{
		Addr:              cfg.HTTPPort,
		Handler:           router.LoggingMiddleware(withCors),
		ReadHeaderTimeout: time.Duration(cfg.ReadTimeout) * time.Second,
	}

	return &App{
		httpServer: gwServer,
		cert:       cfg.Cert,
		key:        cfg.Key,
	}
}

func (app *App) MustRun(ctx context.Context) {
	if err := app.Run(ctx); err != nil {
		logger.Log().Fatal(ctx, "Failed to run http server: %v", err)
	}
}

func (app *App) Run(ctx context.Context) error {
	logger.Log().Info(ctx, "http server started on %s", app.httpServer.Addr)
	return app.httpServer.ListenAndServeTLS(app.cert, app.key)
	// return app.httpServer.ListenAndServe()
}

func (app *App) Stop(ctx context.Context) {
	logger.Log().Info(ctx, "stopping http server")

	if err := app.httpServer.Shutdown(ctx); err != nil {
		logger.Log().Fatal(ctx, "failed to shutdown http server: %w", err)
	}
}
