package model

import (
	audiov1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/audio"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/model/validator"
)

func ValidateUploadRequest(v *validator.Validator, req *audiov1.UploadRequest) {
	v.Check(validator.AtLeast(int(req.GetBeatId()), 1), "beat_id", "must be positive")
	v.Check(validator.AtLeast(int(req.GetBeatmakerId()), 1), "beatmaker_id", "must be positive")
	v.Check(validator.Between(len(req.GetBeatGenre()), 0, 10), "beat_genre", "must have length between 0 and 10")
}
