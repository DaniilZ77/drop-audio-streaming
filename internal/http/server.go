package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	sl "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
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
	log           *slog.Logger
}

func (r *Router) errorResponse(w http.ResponseWriter, err error, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": err.Error()}); err != nil {
		r.log.Error("write error", sl.Err(err))
	}
}

func NewRouter(
	app *runtime.ServeMux,
	beatProvider BeatProvider,
	mediaUploader MediaUploader,
	log *slog.Logger,
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

func parseRangeHeader(req *http.Request) (start, end *int, err error) {
	rng := strings.TrimPrefix(req.Header.Get("Range"), "bytes=")
	if rng == "" {
		return nil, nil, nil
	}

	vals := strings.Split(rng, "-")
	if len(vals) != 2 {
		return nil, nil, fmt.Errorf("%w: must be in form bytes=99-9999", model.ErrInvalidRangeHeader)
	}

	s, err := strconv.Atoi(vals[0])
	if err != nil || s < 0 {
		return nil, nil, fmt.Errorf("%w: must be non negative integer", model.ErrInvalidRangeHeader)
	}

	if vals[1] == "" {
		return &s, nil, nil
	}

	e, err := strconv.Atoi(vals[1])
	if err != nil || e < s {
		return nil, nil, fmt.Errorf("%w: must be positive integer gte start", model.ErrInvalidRangeHeader)
	}

	return &s, &e, nil
}

func (r *Router) stream(w http.ResponseWriter, req *http.Request, params map[string]string) {
	ctx := req.Context()

	data := params["id"]
	if err := uuid.Validate(data); err != nil {
		r.errorResponse(w, err, http.StatusBadRequest)
		return
	}

	beatID, err := uuid.Parse(data)
	if err != nil {
		r.errorResponse(w, err, http.StatusBadRequest)
		return
	}

	s, e, err := parseRangeHeader(req)
	if err != nil {
		if errors.Is(err, model.ErrInvalidRangeHeader) {
			r.errorResponse(w, err, http.StatusBadRequest)
			return
		}
		r.log.Error("internal error", sl.Err(err))
		r.errorResponse(w, err, http.StatusInternalServerError)
		return
	}

	beat, size, contentType, err := r.beatProvider.GetBeatStream(ctx, beatID, s, e)
	if err != nil {
		if errors.Is(err, model.ErrBeatNotFound) || errors.Is(err, model.ErrInvalidRangeHeader) {
			r.errorResponse(w, err, http.StatusBadRequest)
			return
		}
		r.log.Error("internal error", sl.Err(err))
		r.errorResponse(w, err, http.StatusInternalServerError)
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
		r.log.Error("write error", sl.Err(err))
	}
}

func parseUploadParams(req *http.Request) (*model.MediaMeta, error) {
	t := req.URL.Query().Get("type")
	if t != "file" && t != "archive" && t != "image" {
		return nil, &model.ModelError{Err: model.ErrInvalidMediaType}
	}

	exp, err := strconv.ParseInt(req.URL.Query().Get("exp"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: exp must be integer", err)
	}

	return &model.MediaMeta{
		MediaType:         model.MediaType(t),
		HttpContentType:   req.Header.Get("Content-Type"),
		HttpContentLength: req.ContentLength,
		Name:              req.URL.Query().Get("name"),
		Expiry:            exp,
		UploadURL:         req.URL.String(),
	}, nil
}

func (r *Router) upload(w http.ResponseWriter, req *http.Request, params map[string]string) {
	ctx := req.Context()

	m, err := parseUploadParams(req)
	if err != nil {
		r.errorResponse(w, err, http.StatusBadRequest)
		return
	}

	defer req.Body.Close()

	if err := r.mediaUploader.UploadMedia(ctx, req.Body, *m); err != nil {
		var modelErr *model.ModelError
		if errors.As(err, &modelErr) {
			r.errorResponse(w, err, http.StatusBadRequest)
			return
		}
		r.log.Error("internal error", sl.Err(err))
		r.errorResponse(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
