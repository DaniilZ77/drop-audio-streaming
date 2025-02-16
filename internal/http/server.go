package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/model"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type BeatProvider interface {
	GetBeatStream(ctx context.Context, beatID uuid.UUID, start, end *int) (file io.ReadCloser, size *int, contentType *string, err error)
}

type MediaUploader interface {
	UploadMedia(ctx context.Context, file io.Reader, m model.MediaMeta) error
}

type Router struct {
	app           *runtime.ServeMux
	beatProvider  BeatProvider
	mediaUploader MediaUploader
}

func NewRouter(
	app *runtime.ServeMux,
	beatProvider BeatProvider,
	mediaUploader MediaUploader,
) {
	r := &Router{
		app:           app,
		beatProvider:  beatProvider,
		mediaUploader: mediaUploader,
	}

	r.initRoutes()
}

func (r *Router) initRoutes() {
	_ = r.app.HandlePath(http.MethodGet, "/v1/beat/{id}/stream", r.stream)
	_ = r.app.HandlePath(http.MethodPut, "/v1/beat", r.upload)
}

func parseRangeHeader(ctx context.Context, req *http.Request) (start, end *int, err error) {
	rng := strings.TrimPrefix(req.Header.Get("Range"), "bytes=")
	if rng == "" {
		return nil, nil, nil
	}

	vals := strings.Split(rng, "-")
	if len(vals) != 2 {
		logger.Log().Error(ctx, model.ErrInvalidRangeHeader.Error())
		return nil, nil, model.ErrInvalidRangeHeader
	}

	s, err := strconv.Atoi(vals[0])
	if err != nil || s < 0 {
		logger.Log().Error(ctx, model.ErrInvalidRangeHeader.Error())
		return nil, nil, model.ErrInvalidRangeHeader
	}

	if vals[1] == "" {
		return &s, nil, nil
	}

	e, err := strconv.Atoi(vals[1])
	if err != nil || e < s {
		logger.Log().Error(ctx, model.ErrInvalidRangeHeader.Error())
		return nil, nil, model.ErrInvalidRangeHeader
	}

	return &s, &e, nil
}

func (r *Router) stream(w http.ResponseWriter, req *http.Request, params map[string]string) {
	ctx := req.Context()

	data := params["id"]
	if err := uuid.Validate(data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	beatID := uuid.MustParse(data)

	s, e, err := parseRangeHeader(ctx, req)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, model.ErrInvalidRangeHeader) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	beat, size, contentType, err := r.beatProvider.GetBeatStream(ctx, beatID, s, e)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, model.ErrBeatNotFound) || errors.Is(err, model.ErrInvalidRangeHeader) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer beat.Close()

	w.Header().Set("Content-Type", *contentType)
	w.Header().Set("Connection", "keep-alive")

	if s != nil {
		if e == nil || *e >= *size {
			e = new(int)
			*e = *size - 1
		}
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", *s, *e, *size))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", *e-*s+1))
		w.WriteHeader(http.StatusPartialContent)
	} else {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", *size))
		w.WriteHeader(http.StatusOK)
	}

	if _, err = io.Copy(w, beat); err != nil {
		logger.Log().Error(ctx, err.Error())
	}
}

func parseUploadParams(ctx context.Context, req *http.Request) (*model.MediaMeta, error) {
	t := req.URL.Query().Get("type")
	if t != "file" && t != "archive" && t != "image" {
		logger.Log().Debug(ctx, model.ErrInvalidType.Error())
		return nil, &model.ModelError{Err: model.ErrInvalidType}
	}

	exp, err := strconv.Atoi(req.URL.Query().Get("exp"))
	if err != nil {
		logger.Log().Debug(ctx, err.Error())
		return nil, err
	}

	return &model.MediaMeta{
		MediaType:     model.MediaType(t),
		ContentType:   req.Header.Get("Content-Type"),
		ContentLength: req.ContentLength,
		Name:          req.URL.Query().Get("name"),
		Expiry:        int64(exp),
		URL:           req.URL.String(),
	}, nil
}

func (r *Router) upload(w http.ResponseWriter, req *http.Request, params map[string]string) {
	ctx := req.Context()

	m, err := parseUploadParams(ctx, req)
	if err != nil {
		logger.Log().Debug(ctx, err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer req.Body.Close()

	if err := r.mediaUploader.UploadMedia(ctx, req.Body, *m); err != nil {
		logger.Log().Error(ctx, err.Error())
		var modelErr *model.ModelError
		if errors.As(err, &modelErr) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
