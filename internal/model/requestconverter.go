package model

import (
	"net/url"
	"slices"
	"strconv"
	"strings"

	audiov1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/audio"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/model/validator"
)

func ToCoreBeat(req *audiov1.UploadRequest) core.BeatParams {
	slices.Sort(req.BeatGenre)
	req.BeatGenre = slices.Compact(req.BeatGenre)

	slices.Sort(req.BeatTag)
	req.BeatTag = slices.Compact(req.BeatTag)

	slices.Sort(req.BeatMood)
	req.BeatMood = slices.Compact(req.BeatMood)

	var genres []core.BeatGenre
	var tags []core.BeatTag
	var moods []core.BeatMood

	for _, genre := range req.GetBeatGenre() {
		genres = append(genres, core.BeatGenre{
			BeatID:  int(req.GetBeatId()),
			GenreID: int(genre),
		})
	}

	for _, tag := range req.GetBeatTag() {
		tags = append(tags, core.BeatTag{
			BeatID: int(req.GetBeatId()),
			TagID:  int(tag),
		})
	}

	for _, mood := range req.GetBeatMood() {
		moods = append(moods, core.BeatMood{
			BeatID: int(req.GetBeatId()),
			MoodID: int(mood),
		})
	}

	return core.BeatParams{
		Beat: core.Beat{
			ID:          int(req.GetBeatId()),
			BeatmakerID: int(req.GetBeatmakerId()),
			Name:        req.GetName(),
			Description: req.GetDescription(),
			Bpm:         int(req.GetBpm()),
		},
		Genres: genres,
		Tags:   tags,
		Moods:  moods,
		Note: core.BeatNote{
			BeatID: int(req.GetBeatId()),
			NoteID: int(req.GetNote().GetNoteId()),
			Scale:  req.GetNote().GetScale(),
		},
	}
}

func toIntSlice(str []string) ([]int, error) {
	var intSlice []int
	for _, s := range str {
		if s == "" {
			continue
		}

		i, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		intSlice = append(intSlice, i)
	}
	return intSlice, nil
}

func ToCoreFeedFilter(v *validator.Validator, params url.Values) (*core.FeedFilter, error) {
	genresStr := strings.Split(params.Get("genres"), ",")
	genres, err := toIntSlice(genresStr)
	if err != nil {
		v.AddError("genre", "must be integer")
		return nil, core.ErrValidationFailed
	}

	moodsStr := strings.Split(params.Get("moods"), ",")
	moods, err := toIntSlice(moodsStr)
	if err != nil {
		v.AddError("mood", "must be integer")
		return nil, core.ErrValidationFailed
	}

	tagsStr := strings.Split(params.Get("tags"), ",")
	tags, err := toIntSlice(tagsStr)
	if err != nil {
		v.AddError("tag", "must be integer")
		return nil, core.ErrValidationFailed
	}

	noteStr := strings.Split(params.Get("note"), ",")
	if len(noteStr) != 2 && len(noteStr) != 1 {
		v.AddError("note", "must be in form note_id,scale")
		return nil, core.ErrValidationFailed
	}
	var note *struct {
		NoteID int
		Scale  string
	}

	if len(noteStr) == 2 {
		noteID, err := strconv.Atoi(noteStr[0])
		if err != nil {
			v.AddError("note_id", "note_id must be integer")
			return nil, core.ErrValidationFailed
		}

		note = &struct {
			NoteID int
			Scale  string
		}{
			NoteID: noteID,
			Scale:  noteStr[1],
		}
	}

	bpmStr := params.Get("bpm")
	var bpm *int
	if bpmStr != "" {
		localBpm, err := strconv.Atoi(params.Get("bpm"))
		if err != nil {
			v.AddError("bpm", "must be integer")
			return nil, err
		}
		bpm = &localBpm
	}

	return &core.FeedFilter{
		Genres: genres,
		Moods:  moods,
		Tags:   tags,
		Note:   note,
		Bpm:    bpm,
	}, nil
}

func ToGetBeatsParams(limit, offset int, order string) *core.GetBeatsParams {
	return &core.GetBeatsParams{
		Limit:  limit,
		Offset: offset,
		Order:  order,
	}
}
