package model

import (
	audiov1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/audio"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/model/validator"
)

func ValidateUploadRequest(v *validator.Validator, req *audiov1.UploadRequest) {
	v.Check(validator.AtLeast(int(req.GetBeatId()), 1), "beat_id", "must be positive")
	v.Check(validator.AtLeast(int(req.GetBeatmakerId()), 1), "beatmaker_id", "must be positive")
	v.Check(validator.Between(len(req.GetBeatGenre()), 1, 10), "beat_genre", "must have length between 0 and 10")
	v.Check(validator.Between(len(req.GetName()), 1, 80), "name", "must have length between 1 and 80")
	v.Check(validator.Between(len(req.GetDescription()), 0, 512), "description", "must have length between 0 and 512")
}

func ValidateGetBeat(v *validator.Validator, req *audiov1.GetBeatRequest) {
	v.Check(validator.AtLeast(int(req.GetBeatId()), 1), "beat_id", "must be positive")
}

func ValidateGetBeatmakerBeats(v *validator.Validator, getBeatsParams GetBeatsParams, beatmakerID int) {
	v.Check(validator.Between(getBeatsParams.Limit, 1, 100), "limit", "must be positive and less than or equal 100")
	v.Check(validator.AtLeast(getBeatsParams.Offset, 0), "offset", "must be non-negative")
	v.Check(validator.OneOf(getBeatsParams.Order, "asc", "desc"), "order", "must be one of acs or desc")
	v.Check(validator.AtLeast(beatmakerID, 1), "beatmaker_id", "must be positive")
}
