package http

import (
	"context"
	"net/http"
	"time"

	"net/http/pprof"

	audiov1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/audio"
	client "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/client"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/config"
	router "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/http"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
	beat "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/service"
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
	beatService *beat.BeatService,
	grpcUserClient *client.Client,
) *App {
	// creds, err := credentials.NewClientTLSFromFile(cfg.Cert, "") nolint
	// if err != nil {
	// 	logger.Log().Fatal(ctx, "failed to create server TLS credentials: %v", err)
	// }

	conn, err := grpc.NewClient(cfg.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log().Fatal(ctx, "failed to dial server: %s", err.Error())
	}

	gwmux := runtime.NewServeMux()
	router.NewRouter(gwmux, beatService, beatService)

	// Register user
	err = audiov1.RegisterBeatServiceHandler(ctx, gwmux, conn)
	if err != nil {
		logger.Log().Fatal(ctx, "failed to register gateway: %v", err.Error())
	}

	mux := http.NewServeMux()
	mux.Handle("/", gwmux)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)

	// Cors
	withCors := cors.AllowAll().Handler(mux)

	// Server
	gwServer := &http.Server{
		Addr:              cfg.HTTPPort,
		Handler:           interceptorLogger(withCors),
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
		logger.Log().Fatal(ctx, "failed to run http server: %s", err.Error())
	}
}

func (app *App) Run(ctx context.Context) error {
	logger.Log().Info(ctx, "http server started on %s", app.httpServer.Addr)
	// return app.httpServer.ListenAndServeTLS(app.cert, app.key)
	return app.httpServer.ListenAndServe()
}

func (app *App) Stop(ctx context.Context) {
	logger.Log().Info(ctx, "stopping http server")
	if err := app.httpServer.Shutdown(ctx); err != nil {
		logger.Log().Fatal(ctx, "failed to shutdown http server: %s", err.Error())
	}
}

func interceptorLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		logger.Log().Debug(req.Context(), req.URL.String())

		h.ServeHTTP(w, req)
	})
}
