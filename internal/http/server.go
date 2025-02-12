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
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type BeatProvider interface {
	GetBeatStream(ctx context.Context, beatID int, start, end *int) (file io.ReadCloser, size *int, contentType *string, err error)
}

type BeatStreamer interface {
	StreamBeat(ctx context.Context, r io.Reader, w io.Writer, chunkSize int) error
}

type Router struct {
	app          *runtime.ServeMux
	beatProvider BeatProvider
	beatStreamer BeatStreamer
	chunkSize    int
}

func NewRouter(
	app *runtime.ServeMux,
	beatProvider BeatProvider,
	beatStreamer BeatStreamer,
	chunkSize int,
) {
	r := &Router{
		app:          app,
		beatProvider: beatProvider,
		beatStreamer: beatStreamer,
		chunkSize:    chunkSize,
	}

	r.initRoutes()
}

func (r *Router) initRoutes() {
	_ = r.app.HandlePath(http.MethodGet, "/v1/beat/{id}/stream", r.stream)
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

	beatID, err := strconv.Atoi(params["id"])
	if err != nil || beatID < 1 {
		logger.Log().Error(ctx, model.ErrInvalidBeatID.Error())
		http.Error(w, model.ErrInvalidBeatID.Error(), http.StatusBadRequest)
	}

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

	if err = r.beatStreamer.StreamBeat(ctx, beat, w, r.chunkSize); err != nil {
		logger.Log().Error(ctx, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
