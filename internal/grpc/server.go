package grpc

import (
	"context"
	"errors"
	"log/slog"

	audiov1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/audio"
	userv1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/user"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/db/generated"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/domain/model"
	sl "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
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
	GetBeats(ctx context.Context, params model.GetBeatsParams) (beats []model.Beat, total *uint64, err error)
	GetBeatParams(ctx context.Context) (params *model.BeatParams, err error)
}

type URLProvider interface {
	GetBeatArchive(ctx context.Context, params generated.SaveOwnerParams) (*string, error)
}

type UserProvider interface {
	GetUser(ctx context.Context, id uuid.UUID) (*userv1.GetUserResponse, error)
}

type server struct {
	audiov1.UnimplementedBeatServiceServer
	beatModifier BeatModifier
	beatProvider BeatProvider
	urlProvider  URLProvider
	userProvider UserProvider
	log          *slog.Logger
}

func Register(
	gRPCServer *grpc.Server,
	beatSaver BeatModifier,
	beatProvider BeatProvider,
	urlProvider URLProvider,
	userProvider UserProvider,
	log *slog.Logger) {
	audiov1.RegisterBeatServiceServer(gRPCServer, &server{beatModifier: beatSaver, beatProvider: beatProvider, urlProvider: urlProvider, userProvider: userProvider, log: log})
}

func (s *server) GetBeatParams(ctx context.Context, req *audiov1.GetBeatParamsRequest) (*audiov1.GetBeatParamsResponse, error) {
	beat, err := s.beatProvider.GetBeatParams(ctx)
	if err != nil {
		s.log.Error("internal error", sl.Err(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return model.ToGetBeatParamsResponse(*beat), nil
}

func (s *server) GetBeats(ctx context.Context, req *audiov1.GetBeatsRequest) (*audiov1.GetBeatsResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	params, err := model.ToModelGetBeatsParams(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	beats, total, err := s.beatProvider.GetBeats(ctx, *params)
	if err != nil {
		s.log.Error("internal error", sl.Err(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	var users []*userv1.GetUserResponse
	for i := range beats {
		user, err := s.userProvider.GetUser(ctx, beats[i].BeatmakerID)
		if err != nil {
			s.log.Error("internal error", sl.Err(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
		users = append(users, user)
	}

	return model.ToGetBeatsResponse(beats, users, *total, *params), nil
}

func (s *server) UploadBeat(ctx context.Context, req *audiov1.UploadBeatRequest) (*audiov1.UploadBeatResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	beat, err := model.ToModelSaveBeatParams(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	fileUploadURL, imageUploadURL, archiveUploadURL, err := s.beatModifier.SaveBeat(ctx, *beat)
	if err != nil {
		var modelErr *model.ModelError
		if errors.Is(err, model.ErrBeatAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		} else if errors.As(err, &modelErr) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		s.log.Error("internal error", sl.Err(err))
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
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	beatID, err := uuid.Parse(req.BeatId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "beat id must be uuid")
	}

	if err := s.beatModifier.DeleteBeat(ctx, beatID); err != nil {
		s.log.Error("internal error", sl.Err(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &audiov1.DeleteBeatResponse{}, nil
}

func (s *server) UpdateBeat(ctx context.Context, req *audiov1.UpdateBeatRequest) (*audiov1.UpdateBeatResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	params, err := model.ToModelUpdateBeatParams(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	fileUploadURL, imageUploadURL, archiveUploadURL, err := s.beatModifier.UpdateBeat(ctx, *params)
	if err != nil {
		var modelErr *model.ModelError
		if errors.Is(err, model.ErrBeatNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		} else if errors.As(err, &modelErr) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		s.log.Error("internal error", sl.Err(err))
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
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	params, err := model.ToModelGetArchiveParams(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	url, err := s.urlProvider.GetBeatArchive(ctx, *params)
	if err != nil {
		var modelErr *model.ModelError
		if errors.Is(err, model.ErrArchiveNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		} else if errors.As(err, &modelErr) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		s.log.Error("internal error", sl.Err(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &audiov1.AcquireBeatResponse{
		ArchiveDownloadUrl: *url,
	}, nil
}

func (s *server) Health(context.Context, *audiov1.HealthRequest) (*audiov1.HealthResponse, error) {
	return &audiov1.HealthResponse{
		Message: "OK",
	}, nil
}
