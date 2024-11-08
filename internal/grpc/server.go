package grpc

import (
	"context"
	"errors"

	audiov1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/audio"
	userclient "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/client/user/grpc"
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
	userClient  *userclient.Client
}

func Register(gRPCServer *grpc.Server, beatService core.BeatService, userClient *userclient.Client) {
	audiov1.RegisterAudioServiceServer(gRPCServer, &server{beatService: beatService, userClient: userClient})
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

func (s *server) GetBeatMeta(ctx context.Context, req *audiov1.GetBeatMetaRequest) (*audiov1.GetBeatMetaResponse, error) {
	v := validator.New()
	model.ValidateGetBeatMeta(v, req)
	if !v.Valid() {
		logger.Log().Debug(ctx, "validation failed: %v", v.Errors)
		return nil, toGRPCError(v)
	}

	beat, beatGenres, err := s.beatService.GetBeatMeta(ctx, int(req.GetBeatId()))
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, core.ErrBeatNotFound) {
			return nil, status.Error(codes.NotFound, core.ErrBeatNotFound.Error())
		}
		return nil, status.Error(codes.Internal, core.ErrInternal.Error())
	}

	beatmaker, err := s.userClient.GetUserByID(ctx, beat.BeatmakerID)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, status.Error(codes.Internal, core.ErrInternal.Error())
	}

	return model.ToGetBeatMetaResponse(beat, beatmaker, beatGenres), nil
}
