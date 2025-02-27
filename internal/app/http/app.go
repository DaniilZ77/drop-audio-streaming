package http

import (
	"context"
	"log/slog"
	"net/http"

	"net/http/pprof"

	audiov1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/audio"
	client "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/client"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/config"
	router "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/http"
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
	log        *slog.Logger
}

func New(
	ctx context.Context,
	cfg *config.Config,
	beatService *beat.BeatService,
	grpcUserClient *client.Client,
	log *slog.Logger,
) *App {
	// creds, err := credentials.NewClientTLSFromFile(cfg.Cert, "") nolint
	// if err != nil {
	// 	logger.Log().Fatal(ctx, "failed to create server TLS credentials: %v", err)
	// }

	conn, err := grpc.NewClient(cfg.GrpcPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	gwmux := runtime.NewServeMux()
	router.NewRouter(gwmux, beatService, beatService, log)

	// Register user
	err = audiov1.RegisterBeatServiceHandler(ctx, gwmux, conn)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", gwmux)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)

	// Cors
	withCors := cors.AllowAll().Handler(mux)

	// Server
	gwServer := &http.Server{
		Addr:    cfg.HttpPort,
		Handler: interceptorLogger(withCors, log),
	}

	return &App{
		httpServer: gwServer,
		cert:       cfg.Tls.Cert,
		key:        cfg.Tls.Key,
		log:        log,
	}
}

func (app *App) MustRun(ctx context.Context) {
	if err := app.Run(ctx); err != nil {
		panic(err)
	}
}

func (app *App) Run(ctx context.Context) error {
	app.log.Info("http server started", slog.String("port", app.httpServer.Addr))
	// return app.httpServer.ListenAndServeTLS(app.cert, app.key)
	return app.httpServer.ListenAndServe()
}

func (app *App) Stop(ctx context.Context) {
	app.log.Info("stopping http server")
	if err := app.httpServer.Shutdown(ctx); err != nil {
		panic(err)
	}
}

func interceptorLogger(h http.Handler, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.DebugContext(req.Context(), req.URL.String())

		h.ServeHTTP(w, req)
	})
}
