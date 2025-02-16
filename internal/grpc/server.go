package grpc

import (
	"context"
	"errors"

	audiov1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/audio"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/db/generated"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/model"
	"github.com/bufbuild/protovalidate-go"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BeatModifier interface {
	SaveBeat(ctx context.Context, beat model.SaveBeatParams) (fileUploadURL, imageUploadURL, archiveUploadURL *string, err error)
	UpdateBeat(ctx context.Context, beat model.UpdateBeatParams) (fileUploadURL, imageUploadURL, archiveUploadURL *string, err error)
	DeleteBeat(ctx context.Context, id uuid.UUID) error
}

type BeatProvider interface {
	GetBeats(ctx context.Context, params model.GetBeatsParams) (beats []model.Beat, total *int, err error)
	GetBeatParams(ctx context.Context) (params *model.BeatParams, err error)
}

type URLProvider interface {
	GetBeatArchive(ctx context.Context, params generated.SaveOwnerParams) (*string, error)
}

type server struct {
	audiov1.UnimplementedBeatServiceServer
	beatModifier BeatModifier
	beatProvider BeatProvider
	urlProvider  URLProvider
}

func Register(
	gRPCServer *grpc.Server,
	beatSaver BeatModifier,
	beatProvider BeatProvider,
	urlProvider URLProvider) {
	audiov1.RegisterBeatServiceServer(gRPCServer, &server{beatModifier: beatSaver, beatProvider: beatProvider, urlProvider: urlProvider})
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
	fileUploadURL, imageUploadURL, archiveUploadURL, err := s.beatModifier.SaveBeat(ctx, beat)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		var modelErr *model.ModelError
		if errors.Is(err, model.ErrBeatAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		} else if errors.As(err, &modelErr) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &audiov1.UploadBeatResponse{
		FileUploadUrl:    *fileUploadURL,
		ImageUploadUrl:   *imageUploadURL,
		ArchiveUploadUrl: *archiveUploadURL,
	}, nil
}

func (s *server) DeleteBeat(ctx context.Context, req *audiov1.DeleteBeatRequest) (*audiov1.DeleteBeatResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		logger.Log().Debug(ctx, err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := s.beatModifier.DeleteBeat(ctx, uuid.MustParse(req.BeatId)); err != nil {
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
	fileUploadURL, imageUploadURL, archiveUploadURL, err := s.beatModifier.UpdateBeat(ctx, params)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		var modelErr *model.ModelError
		if errors.Is(err, model.ErrBeatNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		} else if errors.As(err, &modelErr) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &audiov1.UpdateBeatResponse{
		FileUploadUrl:    fileUploadURL,
		ImageUploadUrl:   imageUploadURL,
		ArchiveUploadUrl: archiveUploadURL,
	}, nil
}

func (s *server) AcquireBeat(ctx context.Context, req *audiov1.AcquireBeatRequest) (*audiov1.AcquireBeatResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		logger.Log().Debug(ctx, err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	params := model.ToModelGetArchiveParams(req)
	url, err := s.urlProvider.GetBeatArchive(ctx, params)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		var modelErr *model.ModelError
		if errors.Is(err, model.ErrArchiveNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		} else if errors.As(err, &modelErr) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &audiov1.AcquireBeatResponse{
		ArchiveDownloadUrl: *url,
	}, nil
}
