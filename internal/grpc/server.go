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

func (s *server) GetGenres(context.Context, *audiov1.GetGenresRequest) (*audiov1.GetGenresResponse, error) {
	panic("unimplemented")
}

func (s *server) GetMoods(context.Context, *audiov1.GetMoodsRequest) (*audiov1.GetMoodsResponse, error) {
	panic("unimplemented")
}

func (s *server) GetNotes(context.Context, *audiov1.GetNotesRequest) (*audiov1.GetNotesResponse, error) {
	panic("unimplemented")
}

func (s *server) GetTags(context.Context, *audiov1.GetTagsRequest) (*audiov1.GetTagsResponse, error) {
	panic("unimplemented")
}

func Register(gRPCServer *grpc.Server, beatService core.BeatService, userClient *userclient.Client) {
	audiov1.RegisterAudioServiceServer(gRPCServer, &server{beatService: beatService, userClient: userClient})
}

func (s *server) Upload(ctx context.Context, req *audiov1.UploadRequest) (*audiov1.UploadResponse, error) {
	v := validator.New()
	model.ValidateUploadRequest(v, req)
	if !v.Valid() {
		logger.Log().Debug(ctx, "validation failed: %v", v.Errors)
		return nil, withDetails(codes.InvalidArgument, core.ErrValidationFailed, v.Errors)
	}

	beatParams := model.ToCoreBeat(req)

	filePath, imagePath, err := s.beatService.AddBeat(ctx, beatParams)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, core.ErrBeatExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, core.ErrInternal.Error())
	}

	fileURL, err := s.beatService.GetUploadURL(ctx, filePath)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, status.Error(codes.Internal, core.ErrInternal.Error())
	}

	imageURL, err := s.beatService.GetUploadURL(ctx, imagePath)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, status.Error(codes.Internal, core.ErrInternal.Error())
	}

	return model.ToUploadResponse(fileURL, imageURL), nil
}

func (s *server) GetBeat(ctx context.Context, req *audiov1.GetBeatRequest) (*audiov1.GetBeatResponse, error) {
	v := validator.New()
	model.ValidateGetBeat(v, req)
	if !v.Valid() {
		logger.Log().Debug(ctx, "validation failed: %v", v.Errors)
		return nil, withDetails(codes.InvalidArgument, core.ErrValidationFailed, v.Errors)
	}

	beat, err := s.beatService.GetBeat(ctx, int(req.GetBeatId()))
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, core.ErrBeatNotFound) {
			return nil, status.Error(codes.NotFound, core.ErrBeatNotFound.Error())
		}
		return nil, status.Error(codes.Internal, core.ErrInternal.Error())
	}

	beatmaker, err := s.userClient.GetUserByID(ctx, beat.Beat.BeatmakerID)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, status.Error(codes.Internal, core.ErrInternal.Error())
	}

	return model.ToGetBeatResponse(beat, beatmaker), nil
}
