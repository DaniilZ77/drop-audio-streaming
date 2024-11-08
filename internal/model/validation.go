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

func ValidateGetBeatMeta(v *validator.Validator, req *audiov1.GetBeatMetaRequest) {
	v.Check(validator.AtLeast(int(req.GetBeatId()), 1), "beat_id", "must be positive")
}
