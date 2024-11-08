package grpc

import (
	"context"
	"errors"

	audiov1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/audio"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/model"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/model/validator"
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
	v := validator.New()
	model.ValidateUploadRequest(v, req)
	if !v.Valid() {
		logger.Log().Debug(ctx, "validation failed: %v", v.Errors)
		return nil, toGRPCError(v)
	}

	beat := model.ToCoreBeat(req)
	beatGenre := model.ToCoreBeatGenre(req)

	beatPath, err := s.beatService.AddBeat(ctx, beat, beatGenre)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, core.ErrBeatExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, core.ErrInternal.Error())
	}

	url, err := s.beatService.GetUploadURL(ctx, beatPath)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, status.Error(codes.Internal, core.ErrInternal.Error())
	}

	return &audiov1.UploadResponse{BeatUploadUrl: url}, nil
}
