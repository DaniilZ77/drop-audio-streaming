package grpc

import (
	"context"
	"errors"

	audiov1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/audio"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	audiov1.UnimplementedAudioServiceServer
	beatService core.BeatService
}

func Register(gRPCServer *grpc.Server, beatService core.BeatService) {
	audiov1.RegisterAudioServiceServer(gRPCServer, &server{beatService: beatService})
}

func (s *server) Upload(ctx context.Context, req *audiov1.UploadRequest) (*audiov1.UploadResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, core.ErrUnauthorized) {
			return nil, status.Error(codes.Unauthenticated, core.ErrUnauthorized.Error())
		}
		return nil, status.Error(codes.Internal, core.ErrInternal.Error())
	}

	_, beatPath, err := s.beatService.AddBeat(ctx, userID)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, status.Error(codes.Internal, core.ErrInternal.Error())
	}

	url, err := s.beatService.GetUploadURL(ctx, beatPath)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, status.Error(codes.Internal, core.ErrInternal.Error())
	}

	return &audiov1.UploadResponse{Url: url}, nil
}
