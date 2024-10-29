package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
)

func (r *Router) stream(w http.ResponseWriter, req *http.Request, params map[string]string) {
	ctx := req.Context()

	id, err := strconv.ParseInt(params["id"], 10, 64)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		http.Error(w, core.ErrInternal.Error(), http.StatusInternalServerError)
		return
	}

	start, end, err := parseRangeHeader(ctx, req)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, core.ErrInvalidRange) {
			http.Error(w, core.ErrInvalidRange.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, core.ErrInternal.Error(), http.StatusInternalServerError)
		return
	}

	beat, size, err := r.beatService.GetBeat(ctx, id, start, &end)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, core.ErrBeatNotFound) {
			http.Error(w, core.ErrBeatNotFound.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, core.ErrInternal.Error(), http.StatusInternalServerError)
		return
	}
	defer beat.Close()

	w.Header().Set("Content-Type", "audio/mpeg")
	if start != 0 || end != -1 {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, size))
		w.WriteHeader(http.StatusPartialContent)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err = r.beatService.WritePartialContent(ctx, beat, w, r.chunkSize); err != nil {
		logger.Log().Error(ctx, err.Error())
		http.Error(w, core.ErrInternal.Error(), http.StatusInternalServerError)
	}
}

func parseRangeHeader(ctx context.Context, req *http.Request) (start, end int64, err error) {
	val := strings.TrimPrefix(req.Header.Get("Range"), "bytes=")
	if val == "" {
		return 0, -1, nil
	}

	tmp := strings.Split(val, "-")
	if len(tmp) != 2 {
		logger.Log().Error(ctx, "invalid range header")
		return 0, 0, core.ErrInvalidRange
	}

	start, err = strconv.ParseInt(tmp[0], 10, 64)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return 0, 0, core.ErrInvalidRange
	}

	end, err = strconv.ParseInt(tmp[1], 10, 64)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return 0, 0, core.ErrInvalidRange
	}

	if start < 0 || end < start {
		logger.Log().Error(ctx, "invalid range header")
		return 0, 0, core.ErrInvalidRange
	}

	return start, end, nil
}
