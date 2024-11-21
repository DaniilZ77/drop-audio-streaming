package model

import (
	"net/url"
	"strconv"

	audiov1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/audio"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/model/validator"
)

func ValidateUploadRequest(v *validator.Validator, req *audiov1.UploadRequest) {
	v.Check(validator.AtLeast(int(req.GetBeatId()), 1), "beat_id", "must be positive")
	v.Check(validator.AtLeast(int(req.GetBeatmakerId()), 1), "beatmaker_id", "must be positive")
	v.Check(validator.Between(len(req.GetBeatGenre()), 1, 10), "beat_genre", "must have length between 1 and 10")
	v.Check(validator.Between(len(req.GetName()), 1, 80), "name", "must have length between 1 and 80")
	v.Check(validator.Between(len(req.GetDescription()), 0, 512), "description", "must have length between 0 and 512")
}

func ValidateGetBeat(v *validator.Validator, req *audiov1.GetBeatRequest) {
	v.Check(validator.AtLeast(int(req.GetBeatId()), 1), "beat_id", "must be positive")
}

func ValidateGetBeatmakerBeats(
	v *validator.Validator,
	values url.Values,
	idStr string,
	id *int,
	params *core.GetBeatsParams) {
	limitStr := values.Get("limit")
	offsetStr := values.Get("offset")
	order := values.Get("order")

	if !validator.IsInteger(limitStr) {
		v.Check(false, "limit", "must be integer")
	} else {
		params.Limit, _ = strconv.Atoi(limitStr)
		v.Check(validator.Between(params.Limit, 1, 100), "limit", "must be positive and less than or equal 100")
	}

	if !validator.IsInteger(offsetStr) {
		v.Check(false, "offset", "must be integer")
	} else {
		params.Offset, _ = strconv.Atoi(offsetStr)
		v.Check(validator.AtLeast(params.Offset, 0), "offset", "must be non negative")
	}

	if !validator.IsInteger(idStr) {
		v.Check(false, "beatmaker_id", "must be integer")
	} else {
		*id, _ = strconv.Atoi(idStr)
		v.Check(validator.AtLeast(*id, 1), "beatmaker_id", "must be positive")
	}

	v.Check(validator.OneOf(order, "asc", "desc"), "order", "must be one of acs or desc")
}

func ValidateStream(v *validator.Validator, idStr string, id *int) {
	if !validator.IsInteger(idStr) {
		v.Check(false, "id", "must be integer")
	} else {
		*id, _ = strconv.Atoi(idStr)
		v.Check(validator.AtLeast(*id, 1), "id", "must be positive")
	}
}
