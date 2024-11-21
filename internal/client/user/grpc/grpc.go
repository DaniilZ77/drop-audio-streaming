package grpc

import (
	"context"
	"fmt"
	"time"

	userv1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/user"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	api userv1.UserServiceClient
}

func New(
	ctx context.Context,
	addr string,
	timeout time.Duration,
	retriesCount uint,
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
			logging.UnaryClientInterceptor(interceptorLogger(logger.Log()), logOpts...),
			retry.UnaryClientInterceptor(retryOpts...),
		),
	)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	return &Client{
		api: userv1.NewUserServiceClient(cc),
	}, nil
}

func (c *Client) GetUserByID(ctx context.Context, userID int) (*userv1.GetUserResponse, error) {
	resp, err := c.api.GetUser(ctx, &userv1.GetUserRequest{
		UserId: int64(userID),
	})
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	return resp, nil
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
