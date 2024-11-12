package gprc

import (
	"context"
	"fmt"
	"net"

	userclient "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/client/user/grpc"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/config"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core"
	audio "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/grpc"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type App struct {
	gRPCServer *grpc.Server
	port       string
}

func New(
	ctx context.Context,
	cfg *config.Config,
	beatService core.BeatService,
	grpcUserClient *userclient.Client,
) *App {
	opts := []grpc.ServerOption{}

	// Logger
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(
			logging.PayloadReceived,
			logging.PayloadSent,
		),
	}

	// Recovery
	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p any) (err error) {
			logger.Log().Error(ctx, fmt.Sprintf("recovered from panic: %q", p))

			return status.Errorf(codes.Internal, "internal error")
		}),
	}

	opts = append(opts, grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOpts...),
		logging.UnaryServerInterceptor(interceptorLogger(logger.Log()), loggingOpts...),
	))

	// TLS nolint
	// creds, err := credentials.NewServerTLSFromFile(cfg.Cert, cfg.Key)
	// if err != nil {
	// 	logger.Log().Fatal(ctx, "failed to create server TLS credentials: %v", err)
	// }
	opts = append(opts, grpc.Creds(insecure.NewCredentials()))

	// Create gRPC server
	gRPCServer := grpc.NewServer(opts...)

	// Register services
	audio.Register(gRPCServer, beatService, grpcUserClient)

	return &App{
		gRPCServer: gRPCServer,
		port:       cfg.GRPCPort,
	}
}

func interceptorLogger(l logger.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		switch lvl {
		case logging.LevelDebug:
			l.Debug(ctx, msg, fields...)
		case logging.LevelInfo:
			l.Info(ctx, msg, fields...)
		case logging.LevelWarn:
			l.Warn(ctx, msg, fields...)
		case logging.LevelError:
			l.Error(ctx, msg, fields...)
		default:
			logger.Log().Fatal(ctx, fmt.Sprintf("unknown level %v", lvl))
		}
	})
}

func (a *App) MustRun(ctx context.Context) {
	if err := a.Run(ctx); err != nil {
		logger.Log().Fatal(ctx, "failed to run grpc server: %v", err)
	}
}

func (a *App) Run(ctx context.Context) error {
	l, err := net.Listen("tcp", a.port)
	if err != nil {
		return err
	}

	logger.Log().Info(ctx, fmt.Sprintf("grpc server started on port %s", a.port))

	if err := a.gRPCServer.Serve(l); err != nil {
		return err
	}

	return nil
}

func (a *App) Stop(ctx context.Context) {
	logger.Log().Info(ctx, "stopping grpc server")

	a.gRPCServer.GracefulStop()
}
