package grpc

import (
	"context"
	"errors"

	audiov1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/audio"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/model"
	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BeatModifier interface {
	SaveBeat(ctx context.Context, beat model.SaveBeatParams) (fileUploadURL, imageUploadURL *string, err error)
	UpdateBeat(ctx context.Context, beat model.UpdateBeatParams) (fileUploadURL, imageUploadURL *string, err error)
	DeleteBeat(ctx context.Context, id int) error
}

type BeatProvider interface {
	GetBeats(ctx context.Context, params model.GetBeatsParams) (beats []model.Beat, total *int, err error)
	GetBeatParams(ctx context.Context) (params *model.BeatParams, err error)
}

type server struct {
	audiov1.UnimplementedBeatServiceServer
	beatModifier BeatModifier
	beatProvider BeatProvider
}

func Register(
	gRPCServer *grpc.Server,
	beatSaver BeatModifier,
	beatProvider BeatProvider) {
	audiov1.RegisterBeatServiceServer(gRPCServer, &server{beatModifier: beatSaver, beatProvider: beatProvider})
}

func (s *server) GetBeatParams(ctx context.Context, req *audiov1.GetBeatParamsRequest) (*audiov1.GetBeatParamsResponse, error) {
	beat, err := s.beatProvider.GetBeatParams(ctx)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return model.ToGetBeatParamsResponse(*beat), nil
}

func (s *server) GetBeats(ctx context.Context, req *audiov1.GetBeatsRequest) (*audiov1.GetBeatsResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		logger.Log().Debug(ctx, err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	params := model.ToModelGetBeatsParams(req)
	beats, total, err := s.beatProvider.GetBeats(ctx, params)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, model.ErrOrderByInvalidField) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return model.ToGetBeatsResponse(beats, *total, params), nil
}

func (s *server) UploadBeat(ctx context.Context, req *audiov1.UploadBeatRequest) (*audiov1.UploadBeatResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		logger.Log().Debug(ctx, err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	beat := model.ToModelSaveBeatParams(req)
	fileUploadURL, imageUploadURL, err := s.beatModifier.SaveBeat(ctx, beat)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, model.ErrInvalidGenreID) ||
			errors.Is(err, model.ErrInvalidTagID) ||
			errors.Is(err, model.ErrInvalidMoodID) ||
			errors.Is(err, model.ErrInvalidNoteID) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		} else if errors.Is(err, model.ErrBeatAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &audiov1.UploadBeatResponse{
		FileUploadUrl:  *fileUploadURL,
		ImageUploadUrl: *imageUploadURL,
	}, nil
}

func (s *server) DeleteBeat(ctx context.Context, req *audiov1.DeleteBeatRequest) (*audiov1.DeleteBeatResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		logger.Log().Debug(ctx, err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := s.beatModifier.DeleteBeat(ctx, int(req.BeatId)); err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &audiov1.DeleteBeatResponse{}, nil
}

func (s *server) UpdateBeat(ctx context.Context, req *audiov1.UpdateBeatRequest) (*audiov1.UpdateBeatResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		logger.Log().Debug(ctx, err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	params := model.ToModelUpdateBeatParams(req)
	fileUploadURL, imageUploadURL, err := s.beatModifier.UpdateBeat(ctx, params)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, model.ErrInvalidGenreID) ||
			errors.Is(err, model.ErrInvalidTagID) ||
			errors.Is(err, model.ErrInvalidMoodID) ||
			errors.Is(err, model.ErrInvalidNoteID) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		} else if errors.Is(err, model.ErrBeatNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &audiov1.UpdateBeatResponse{
		FileUploadUrl:  fileUploadURL,
		ImageUploadUrl: imageUploadURL,
	}, nil
}
