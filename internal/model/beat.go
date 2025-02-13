package model

import (
	"strings"
	"time"

	audiov1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/audio"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/db/generated"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type (
	Beat struct {
		ID                int
		BeatmakerID       int
		FilePath          string
		ImagePath         string
		Name              string
		Description       string
		IsFileDownloaded  bool
		IsImageDownloaded bool
		Bpm               int
		CreatedAt         time.Time
		Genres            []string
		Tags              []string
		Moods             []string
		NoteName          *string
		NoteScale         *string
	}

	GetBeatsNote struct {
		Name  string
		Scale string
	}

	OrderBy struct {
		Order string
		Field string
	}

	GetBeatsParams struct {
		BeatID       *int
		Genre        []string
		Mood         []string
		Tag          []string
		Note         *GetBeatsNote
		BeatmakerID  *int
		BeatName     *string
		Bpm          *int
		OrderBy      *OrderBy
		Limit        int
		Offset       int
		IsDownloaded *bool
	}

	BeatParams struct {
		Genres []generated.Genre
		Tags   []generated.Tag
		Moods  []generated.Mood
		Notes  []generated.Note
	}

	UpdateBeatParams struct {
		generated.UpdateBeatParams
		Note   *generated.SaveNoteParams
		Genres []generated.SaveGenresParams
		Tags   []generated.SaveTagsParams
		Moods  []generated.SaveMoodsParams
	}

	SaveBeatParams struct {
		generated.SaveBeatParams
		Note   generated.SaveNoteParams
		Genres []generated.SaveGenresParams
		Tags   []generated.SaveTagsParams
		Moods  []generated.SaveMoodsParams
	}
)

func ToModelSaveBeatParams(beat *audiov1.UploadBeatRequest) SaveBeatParams {
	genres := make([]generated.SaveGenresParams, 0, len(beat.BeatGenre))
	tags := make([]generated.SaveTagsParams, 0, len(beat.BeatTag))
	moods := make([]generated.SaveMoodsParams, 0, len(beat.BeatMood))

	for _, v := range beat.BeatGenre {
		genres = append(genres, generated.SaveGenresParams{
			GenreID: int32(v),
			BeatID:  int32(beat.BeatId),
		})
	}

	for _, v := range beat.BeatTag {
		tags = append(tags, generated.SaveTagsParams{
			TagID:  int32(v),
			BeatID: int32(beat.BeatId),
		})
	}

	for _, v := range beat.BeatMood {
		moods = append(moods, generated.SaveMoodsParams{
			MoodID: int32(v),
			BeatID: int32(beat.BeatId),
		})
	}

	return SaveBeatParams{
		SaveBeatParams: generated.SaveBeatParams{
			ID:          int32(beat.BeatId),
			BeatmakerID: int32(beat.BeatmakerId),
			Name:        beat.Name,
			Description: beat.Description,
			Bpm:         int32(beat.Bpm),
		},
		Genres: genres,
		Tags:   tags,
		Moods:  moods,
		Note: generated.SaveNoteParams{
			BeatID: int32(beat.BeatId),
			NoteID: int32(beat.Note.NoteId),
			Scale:  generated.Scale(strings.ToLower(beat.Note.Scale.String())),
		},
	}
}

func ToModelGetBeatsParams(params *audiov1.GetBeatsRequest) GetBeatsParams {
	var res GetBeatsParams

	if params.BeatId != nil {
		beatID := int(*params.BeatId)
		res.BeatID = &beatID
	}

	res.Genre = params.Genre
	res.Mood = params.Mood
	res.Tag = params.Tag

	if params.Note != nil {
		res.Note = &GetBeatsNote{
			Name:  params.Note.Name,
			Scale: strings.ToLower(params.Note.Scale.String()),
		}
	}

	if params.BeatmakerId != nil {
		beatmakerID := int(*params.BeatmakerId)
		res.BeatmakerID = &beatmakerID
	}

	res.BeatName = params.BeatName

	if params.Bpm != nil {
		bpm := int(*params.Bpm)
		res.Bpm = &bpm
	}

	if params.OrderBy != nil {
		res.OrderBy = &OrderBy{
			Order: params.OrderBy.Order.String(),
			Field: params.OrderBy.Field,
		}
	}

	res.Limit = int(params.Limit)
	res.Offset = int(params.Offset)

	res.IsDownloaded = params.IsDownloaded

	return res
}

func ToGetBeatsResponse(beats []Beat, total int, params GetBeatsParams) *audiov1.GetBeatsResponse {
	var res []*audiov1.Beat
	for _, b := range beats {
		res = append(res, toResponseBeat(b))
	}

	return &audiov1.GetBeatsResponse{
		Beats: res,
		Pagination: &audiov1.Pagination{
			Records:        int64(total),
			RecordsPerPage: int64(params.Limit),
			Pages:          (int64(total) + int64(params.Limit) - 1) / int64(params.Limit),
			CurPage:        int64(params.Offset)/int64(params.Limit) + 1,
		},
	}
}

func toResponseBeat(b Beat) *audiov1.Beat {
	scale := audiov1.Scale_MINOR
	major := strings.ToLower(audiov1.Scale_name[int32(audiov1.Scale_MAJOR)])
	if b.NoteScale != nil && *b.NoteScale == major {
		scale = audiov1.Scale_MAJOR
	}

	var name string
	if b.NoteName != nil {
		name = *b.NoteName
	}

	return &audiov1.Beat{
		BeatId: int64(b.ID),
		Beatmaker: &audiov1.Beatmaker{
			Id: int64(b.BeatmakerID),
		},
		Image:       b.ImagePath,
		Name:        b.Name,
		Description: b.Description,
		Genre:       b.Genres,
		Tag:         b.Tags,
		Mood:        b.Moods,
		Note: &audiov1.GetBeatsNote{
			Name:  name,
			Scale: scale,
		},
		Bpm:       int64(b.Bpm),
		CreatedAt: timestamppb.New(b.CreatedAt),
	}
}

func ToGetBeatParamsResponse(b BeatParams) *audiov1.GetBeatParamsResponse {
	genres := make([]*audiov1.GenreParam, 0, len(b.Genres))
	tags := make([]*audiov1.TagParam, 0, len(b.Tags))
	moods := make([]*audiov1.MoodParam, 0, len(b.Moods))
	notes := make([]*audiov1.NoteParam, 0, len(b.Notes))

	for _, v := range b.Genres {
		genres = append(genres, &audiov1.GenreParam{
			GenreId: int64(v.ID),
			Name:    v.Name,
		})
	}

	for _, v := range b.Tags {
		tags = append(tags, &audiov1.TagParam{
			TagId: int64(v.ID),
			Name:  v.Name,
		})
	}

	for _, v := range b.Moods {
		moods = append(moods, &audiov1.MoodParam{
			MoodId: int64(v.ID),
			Name:   v.Name,
		})
	}

	for _, v := range b.Notes {
		notes = append(notes, &audiov1.NoteParam{
			NoteId: int64(v.ID),
			Name:   v.Name,
		})
	}

	return &audiov1.GetBeatParamsResponse{
		Genres: genres,
		Tags:   tags,
		Moods:  moods,
		Notes:  notes,
	}
}

func ToModelUpdateBeatParams(req *audiov1.UpdateBeatRequest) UpdateBeatParams {
	var genres []generated.SaveGenresParams
	var tags []generated.SaveTagsParams
	var moods []generated.SaveMoodsParams

	for _, v := range req.BeatGenre {
		genres = append(genres, generated.SaveGenresParams{
			GenreID: int32(v),
			BeatID:  int32(req.BeatId),
		})
	}

	for _, v := range req.BeatTag {
		tags = append(tags, generated.SaveTagsParams{
			TagID:  int32(v),
			BeatID: int32(req.BeatId),
		})
	}

	for _, v := range req.BeatMood {
		moods = append(moods, generated.SaveMoodsParams{
			MoodID: int32(v),
			BeatID: int32(req.BeatId),
		})
	}

	var note *generated.SaveNoteParams
	if req.Note != nil {
		note = &generated.SaveNoteParams{
			BeatID: int32(req.BeatId),
			NoteID: int32(req.Note.NoteId),
			Scale:  generated.Scale(strings.ToLower(req.Note.Scale.String())),
		}
	}

	var name *string
	if req.Name != nil {
		name = req.Name
	}

	var description *string
	if req.Description != nil {
		description = req.Description
	}

	var bpm *int32
	if req.Bpm != nil {
		tmp := int32(*req.Bpm)
		bpm = &tmp
	}

	var isImageDownloaded *bool
	if req.UpdateImage != nil && *req.UpdateImage {
		tmp := false
		isImageDownloaded = &tmp
	}

	var isFileDownloaded *bool
	if req.UpdateFile != nil && *req.UpdateFile {
		tmp := false
		isFileDownloaded = &tmp
	}

	return UpdateBeatParams{
		UpdateBeatParams: generated.UpdateBeatParams{
			ID:                int32(req.BeatId),
			Name:              name,
			Description:       description,
			Bpm:               bpm,
			IsImageDownloaded: isImageDownloaded,
			IsFileDownloaded:  isFileDownloaded,
		},
		Genres: genres,
		Moods:  moods,
		Tags:   tags,
		Note:   note,
	}
}
