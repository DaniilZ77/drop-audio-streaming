package grpc

import (
	"context"
	"log/slog"
	"time"

	userv1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/user"
	sl "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	api userv1.UserServiceClient
	log *slog.Logger
}

func NewUserClient(ctx context.Context,
	addr string,
	timeout time.Duration,
	retriesCount uint,
	log *slog.Logger,
) (*Client, error) {
	retryOpts := []retry.CallOption{
		retry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
		retry.WithMax(retriesCount),
		retry.WithPerRetryTimeout(timeout),
	}

	logOpts := []logging.Option{
		logging.WithLogOnEvents(logging.PayloadReceived, logging.PayloadSent),
	}

	cc, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			logging.UnaryClientInterceptor(interceptorLogger(log), logOpts...),
			retry.UnaryClientInterceptor(retryOpts...),
		),
	)
	if err != nil {
		panic(err)
	}

	return &Client{
		api: userv1.NewUserServiceClient(cc),
		log: log,
	}, nil
}

func interceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		switch lvl {
		case logging.LevelDebug:
			l.DebugContext(ctx, msg, fields...)
		case logging.LevelInfo:
			l.InfoContext(ctx, msg, fields...)
		case logging.LevelWarn:
			l.WarnContext(ctx, msg, fields...)
		case logging.LevelError:
			l.ErrorContext(ctx, msg, fields...)
		default:
			l.Debug("unknown level", slog.Any("level", lvl))
			panic("unknown level")
		}
	})
}

func (c *Client) Health(ctx context.Context) error {
	_, err := c.api.Health(ctx, &userv1.HealthRequest{})
	return err
}

func (c *Client) GetUser(ctx context.Context, id uuid.UUID) (*userv1.GetUserResponse, error) {
	user, err := c.api.GetUser(ctx, &userv1.GetUserRequest{UserId: id.String()})
	if err != nil {
		c.log.Error("failed to get user", sl.Err(err))
		return nil, err
	}

	return user, err
}
