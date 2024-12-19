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
	v.Check(validator.AtLeast(len(req.GetBeatGenre()), 1), "beat_genre", "must not be empty")
	v.Check(validator.AtLeast(int(req.GetBpm()), 1), "bpm", "must be positive")
	v.Check(validator.AtLeast(len(req.GetBeatTag()), 1), "beat_tag", "must not be empty")
	v.Check(validator.AtLeast(len(req.GetBeatMood()), 1), "beat_mood", "must not be empty")
	v.Check(validator.AtLeast(int(req.GetNote().GetNoteId()), 1), "note_id", "must be positive")
	v.Check(validator.OneOf(req.GetNote().GetScale(), "minor", "major"), "scale", "must be one of minor major")
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
	params.Order = values.Get("order")

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

	v.Check(validator.OneOf(params.Order, "asc", "desc"), "order", "must be one of acs or desc")
}

func ValidateStream(v *validator.Validator, idStr string, id *int) {
	if !validator.IsInteger(idStr) {
		v.Check(false, "id", "must be integer")
	} else {
		*id, _ = strconv.Atoi(idStr)
		v.Check(validator.AtLeast(*id, 1), "id", "must be positive")
	}
}

func ValidateGetBeatFiltered(v *validator.Validator, filters core.FeedFilter) {
	for i := range filters.Genres {
		v.Check(validator.AtLeast(filters.Genres[i], 1), "genre_id", "must be positive")
	}

	for i := range filters.Moods {
		v.Check(validator.AtLeast(filters.Moods[i], 1), "mood_id", "must be positive")
	}

	for i := range filters.Tags {
		v.Check(validator.AtLeast(filters.Tags[i], 1), "tag_id", "must be positive")
	}

	if filters.Note != nil {
		v.Check(validator.AtLeast(filters.Note.NoteID, 1), "note_id", "must be positive")
		v.Check(validator.OneOf(filters.Note.Scale, "minor", "major"), "scale", "must be one of minor or major")
	}

	if filters.Bpm != nil {
		v.Check(validator.AtLeast(*filters.Bpm, 1), "bpm", "must be positive")
	}
}
