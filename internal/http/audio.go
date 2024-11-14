package http

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/model"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/model/validator"
)

func (r *Router) stream(w http.ResponseWriter, req *http.Request, params map[string]string) {
	ctx := req.Context()

	id, err := strconv.ParseInt(params["id"], 10, 64)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		errorResponse(ctx, w, http.StatusBadRequest, core.ErrInvalidID, nil)
		return
	} else if id < 1 {
		logger.Log().Error(ctx, "id should be positive")
		errorResponse(ctx, w, http.StatusBadRequest, core.ErrInvalidID, nil)
		return
	}

	start, end, err := parseRangeHeader(ctx, req)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, core.ErrInvalidRange) {
			errorResponse(ctx, w, http.StatusBadRequest, core.ErrInvalidRange, nil)
			return
		}
		errorResponse(ctx, w, http.StatusInternalServerError, core.ErrInternal, nil)
		return
	}

	logger.Log().Debug(ctx, "start: %d; end: %d", start, end)

	beat, size, contentType, err := r.beatService.GetBeatFromS3(ctx, id, start, &end)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, core.ErrBeatNotFound) {
			errorResponse(ctx, w, http.StatusNotFound, core.ErrBeatNotFound, nil)
			return
		}
		errorResponse(ctx, w, http.StatusInternalServerError, core.ErrInternal, nil)
		return
	}
	defer beat.Close()

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Connection", "keep-alive")
	if start != 0 || end != -1 {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, size))
		w.WriteHeader(http.StatusPartialContent)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err = r.beatService.WritePartialContent(ctx, beat, w, r.chunkSize); err != nil {
		logger.Log().Error(ctx, err.Error())
		errorResponse(ctx, w, http.StatusInternalServerError, core.ErrInternal, nil)
	}
}

func (r *Router) getBeat(w http.ResponseWriter, req *http.Request, params map[string]string) {
	ctx := req.Context()

	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		errorResponse(ctx, w, http.StatusInternalServerError, core.ErrInternal, nil)
		return
	}

	feedFilter := model.ToCoreFeedFilter(req.URL.Query())
	beat, genre, err := r.beatService.GetBeatByFilter(ctx, userID, feedFilter)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, core.ErrBeatNotFound) {
			errorResponse(ctx, w, http.StatusNotFound, core.ErrBeatNotFound, nil)
			return
		}
		errorResponse(ctx, w, http.StatusInternalServerError, core.ErrInternal, nil)
		return
	}

	beatmaker, err := r.userClient.GetUserByID(ctx, beat.BeatmakerID)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		errorResponse(ctx, w, http.StatusInternalServerError, core.ErrInternal, nil)
		return
	}

	apiBeatmaker := model.ToBeatmaker(beatmaker)

	b, err := toJSON(model.ToBeat(beat, apiBeatmaker, *genre))
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		errorResponse(ctx, w, http.StatusServiceUnavailable, core.ErrUnavailable, nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err = w.Write(b); err != nil {
		logger.Log().Error(ctx, err.Error())
		errorResponse(ctx, w, http.StatusInternalServerError, core.ErrInternal, nil)
	}
}

func (r *Router) getBeatmakerBeats(w http.ResponseWriter, req *http.Request, params map[string]string) {
	ctx := req.Context()

	var getBeatsParams model.GetBeatsParams
	if err := decoder.Decode(&getBeatsParams, req.URL.Query()); err != nil {
		logger.Log().Error(ctx, err.Error())
		errorResponse(ctx, w, http.StatusInternalServerError, core.ErrInternal, nil)
		return
	}

	id, err := strconv.ParseInt(params["id"], 10, 64)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		errorResponse(ctx, w, http.StatusBadRequest, core.ErrInvalidID, nil)
		return
	}

	v := validator.New()
	model.ValidateGetBeatmakerBeats(v, getBeatsParams, int(id))
	if !v.Valid() {
		logger.Log().Debug(ctx, "%+v", v.Errors)
		errorResponse(ctx, w, http.StatusBadRequest, core.ErrValidationFailed, []interface{}{model.ToValidationErrors(v)})
		return
	}

	coreGetBeatParams := getBeatsParams.ToCoreGetBeatsParams()

	beats, beatsGenres, total, err := r.beatService.GetBeatsByBeatmakerID(ctx, int(id), coreGetBeatParams)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		errorResponse(ctx, w, http.StatusInternalServerError, core.ErrInternal, nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	b, err := toJSON(model.ToGetBeatmakerBeatsResponse(beats, beatsGenres, coreGetBeatParams, total))
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		errorResponse(ctx, w, http.StatusServiceUnavailable, core.ErrUnavailable, nil)
		return
	}

	if _, err = w.Write(b); err != nil {
		logger.Log().Error(ctx, err.Error())
		errorResponse(ctx, w, http.StatusInternalServerError, core.ErrInternal, nil)
	}
}
