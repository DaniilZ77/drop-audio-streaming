package model

import (
	"strings"
	"time"

	audiov1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/audio"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/db/generated"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type (
	Beat struct {
		ID                  uuid.UUID
		BeatmakerID         uuid.UUID
		FilePath            string
		ImagePath           string
		Name                string
		Description         string
		IsFileDownloaded    bool
		IsImageDownloaded   bool
		IsArchiveDownloaded bool
		Bpm                 int
		RangeStart          int
		RangeEnd            int
		CreatedAt           time.Time
		Genres              []string
		Tags                []string
		Moods               []string
		NoteName            *string
		NoteScale           *string
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
		BeatID       *uuid.UUID
		Genre        []string
		Mood         []string
		Tag          []string
		Note         *GetBeatsNote
		BeatmakerID  *uuid.UUID
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

	MediaType string

	AdminScale string

	MediaMeta struct {
		MediaType     MediaType
		ContentType   string
		ContentLength int64
		Name          string
		Expiry        int64
		URL           string
	}
)

const (
	AdminScaleMinor AdminScale = "minor"
	AdminScaleMajor AdminScale = "major"
	File            MediaType  = "file"
	Archive         MediaType  = "archive"
	Image           MediaType  = "image"
)

func ToModelSaveBeatParams(beat *audiov1.UploadBeatRequest) SaveBeatParams {
	genres := make([]generated.SaveGenresParams, 0, len(beat.BeatGenre))
	tags := make([]generated.SaveTagsParams, 0, len(beat.BeatTag))
	moods := make([]generated.SaveMoodsParams, 0, len(beat.BeatMood))

	for _, v := range beat.BeatGenre {
		genres = append(genres, generated.SaveGenresParams{
			GenreID: uuid.MustParse(v),
			BeatID:  uuid.MustParse(beat.BeatId),
		})
	}

	for _, v := range beat.BeatTag {
		tags = append(tags, generated.SaveTagsParams{
			TagID:  uuid.MustParse(v),
			BeatID: uuid.MustParse(beat.BeatId),
		})
	}

	for _, v := range beat.BeatMood {
		moods = append(moods, generated.SaveMoodsParams{
			MoodID: uuid.MustParse(v),
			BeatID: uuid.MustParse(beat.BeatId),
		})
	}

	return SaveBeatParams{
		SaveBeatParams: generated.SaveBeatParams{
			ID:          uuid.MustParse(beat.BeatId),
			BeatmakerID: uuid.MustParse(beat.BeatmakerId),
			Name:        beat.Name,
			Description: beat.Description,
			Bpm:         int32(beat.Bpm),
			RangeStart:  beat.Range.Start,
			RangeEnd:    beat.Range.End,
		},
		Genres: genres,
		Tags:   tags,
		Moods:  moods,
		Note: generated.SaveNoteParams{
			BeatID: uuid.MustParse(beat.BeatId),
			NoteID: uuid.MustParse(beat.Note.NoteId),
			Scale:  generated.NoteScale(strings.ToLower(beat.Note.Scale.String())),
		},
	}
}

func ToModelGetBeatsParams(params *audiov1.GetBeatsRequest) GetBeatsParams {
	var res GetBeatsParams

	if params.BeatId != nil {
		beatID := uuid.MustParse(*params.BeatId)
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
		beatmakerID := uuid.MustParse(*params.BeatmakerId)
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
		BeatId: b.ID.String(),
		Beatmaker: &audiov1.Beatmaker{
			Id: b.BeatmakerID.String(),
		},
		ImageDownloadUrl: b.ImagePath,
		Name:             b.Name,
		Description:      b.Description,
		Genre:            b.Genres,
		Tag:              b.Tags,
		Mood:             b.Moods,
		Note: &audiov1.GetBeatsNote{
			Name:  name,
			Scale: scale,
		},
		Bpm: int64(b.Bpm),
		Range: &audiov1.Range{
			Start: int64(b.RangeStart),
			End:   int64(b.RangeEnd),
		},
		IsFileUploaded:    b.IsFileDownloaded,
		IsImageUploaded:   b.IsImageDownloaded,
		IsArchiveUploaded: b.IsArchiveDownloaded,
		CreatedAt:         timestamppb.New(b.CreatedAt),
	}
}

func ToGetBeatParamsResponse(b BeatParams) *audiov1.GetBeatParamsResponse {
	genres := make([]*audiov1.GenreParam, 0, len(b.Genres))
	tags := make([]*audiov1.TagParam, 0, len(b.Tags))
	moods := make([]*audiov1.MoodParam, 0, len(b.Moods))
	notes := make([]*audiov1.NoteParam, 0, len(b.Notes))

	for _, v := range b.Genres {
		genres = append(genres, &audiov1.GenreParam{
			GenreId: v.ID.String(),
			Name:    v.Name,
		})
	}

	for _, v := range b.Tags {
		tags = append(tags, &audiov1.TagParam{
			TagId: v.ID.String(),
			Name:  v.Name,
		})
	}

	for _, v := range b.Moods {
		moods = append(moods, &audiov1.MoodParam{
			MoodId: v.ID.String(),
			Name:   v.Name,
		})
	}

	for _, v := range b.Notes {
		notes = append(notes, &audiov1.NoteParam{
			NoteId: v.ID.String(),
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
			GenreID: uuid.MustParse(v),
			BeatID:  uuid.MustParse(req.BeatId),
		})
	}

	for _, v := range req.BeatTag {
		tags = append(tags, generated.SaveTagsParams{
			TagID:  uuid.MustParse(v),
			BeatID: uuid.MustParse(req.BeatId),
		})
	}

	for _, v := range req.BeatMood {
		moods = append(moods, generated.SaveMoodsParams{
			MoodID: uuid.MustParse(v),
			BeatID: uuid.MustParse(req.BeatId),
		})
	}

	var note *generated.SaveNoteParams
	if req.Note != nil {
		note = &generated.SaveNoteParams{
			BeatID: uuid.MustParse(req.BeatId),
			NoteID: uuid.MustParse(req.Note.NoteId),
			Scale:  generated.NoteScale(strings.ToLower(req.Note.Scale.String())),
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

	var isArchiveDownloaded *bool
	if req.UpdateArchive != nil && *req.UpdateArchive {
		tmp := false
		isArchiveDownloaded = &tmp
	}

	var rangeStart, rangeEnd *int64
	if req.Range != nil {
		rangeStart = &req.Range.Start
		rangeEnd = &req.Range.End
	}

	return UpdateBeatParams{
		UpdateBeatParams: generated.UpdateBeatParams{
			ID:                  uuid.MustParse(req.BeatId),
			Name:                name,
			Description:         description,
			Bpm:                 bpm,
			IsImageDownloaded:   isImageDownloaded,
			IsFileDownloaded:    isFileDownloaded,
			IsArchiveDownloaded: isArchiveDownloaded,
			RangeStart:          rangeStart,
			RangeEnd:            rangeEnd,
		},
		Genres: genres,
		Moods:  moods,
		Tags:   tags,
		Note:   note,
	}
}

func ToModelGetArchiveParams(req *audiov1.AcquireBeatRequest) generated.SaveOwnerParams {
	return generated.SaveOwnerParams{
		BeatID: uuid.MustParse(req.BeatId),
		UserID: uuid.MustParse(req.UserId),
	}
}
