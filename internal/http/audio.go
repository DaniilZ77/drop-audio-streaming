package http

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/model"
)

const (
	upperLimit = 100
)

func (r *Router) stream(w http.ResponseWriter, req *http.Request, params map[string]string) {
	ctx := req.Context()

	id, err := strconv.ParseInt(params["id"], 10, 64)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		http.Error(w, core.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	} else if id < 1 {
		logger.Log().Error(ctx, "id should be positive")
		http.Error(w, core.ErrInvalidParams.Error(), http.StatusBadRequest)
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

	logger.Log().Debug(ctx, "start: %d; end: %d", start, end)

	beat, size, contentType, err := r.beatService.GetBeat(ctx, id, start, &end)
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

	w.Header().Set("Content-Type", contentType)
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

func (r *Router) getBeat(w http.ResponseWriter, req *http.Request, params map[string]string) {
	ctx := req.Context()

	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		http.Error(w, core.ErrInternal.Error(), http.StatusInternalServerError)
		return
	}

	beatParams := model.ToCoreBeatParams(req.URL.Query())
	beat, genre, err := r.beatService.GetBeatByParams(ctx, userID, beatParams)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, core.ErrBeatNotFound) {
			http.Error(w, core.ErrBeatNotFound.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, core.ErrInternal.Error(), http.StatusInternalServerError)
		return
	}

	beatmaker, err := r.userClient.GetUserByID(ctx, beat.BeatmakerID)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		http.Error(w, core.ErrInternal.Error(), http.StatusInternalServerError)
		return
	}

	apiBeatmaker := model.ToBeatmaker(beatmaker)

	b, err := toJSON(model.ToBeat(beat, apiBeatmaker, *genre))
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		http.Error(w, core.ErrUnavailable.Error(), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err = w.Write(b); err != nil {
		logger.Log().Error(ctx, err.Error())
		http.Error(w, core.ErrInternal.Error(), http.StatusInternalServerError)
	}
}

func (r *Router) getBeatmakerBeats(w http.ResponseWriter, req *http.Request, params map[string]string) {
	ctx := req.Context()

	limit, err := strconv.ParseInt(req.URL.Query().Get("limit"), 10, 64)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		http.Error(w, core.ErrInvalidLimit.Error(), http.StatusBadRequest)
		return
	} else if limit > upperLimit || limit < 0 {
		logger.Log().Debug(ctx, "limit should be less or equal than 100 and greater or equal than 0")
		http.Error(w, core.ErrInvalidLimit.Error(), http.StatusBadRequest)
		return
	}

	offset, err := strconv.ParseInt(req.URL.Query().Get("offset"), 10, 64)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		http.Error(w, core.ErrInvalidOffset.Error(), http.StatusBadRequest)
		return
	} else if offset < 0 {
		logger.Log().Error(ctx, "offset should be greater or equal than 0")
		http.Error(w, core.ErrInvalidOffset.Error(), http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(params["id"], 10, 64)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		http.Error(w, core.ErrInvalidID.Error(), http.StatusBadRequest)
		return
	} else if id < 1 {
		logger.Log().Error(ctx, "id should be greater than 0")
		http.Error(w, core.ErrInvalidID.Error(), http.StatusBadRequest)
		return
	}

	p := core.Pagination{
		Limit:  int(limit),
		Offset: int(offset),
	}

	beats, beatsGenres, total, err := r.beatService.GetBeatsMetaByBeatmakerID(ctx, int(id), p)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		http.Error(w, core.ErrInternal.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	b, err := toJSON(model.ToGetBeatmakerBeatsResponse(beats, beatsGenres, p, total))
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		http.Error(w, core.ErrUnavailable.Error(), http.StatusServiceUnavailable)
		return
	}

	if _, err = w.Write(b); err != nil {
		logger.Log().Error(ctx, err.Error())
		http.Error(w, core.ErrInternal.Error(), http.StatusInternalServerError)
	}
}
